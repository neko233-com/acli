package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var wordCmd = &cobra.Command{Use: "word", Aliases: []string{"docx"}, Short: "DOCX text CRUD operations"}

var wordCreateCmd = &cobra.Command{
	Use: "create <file> <text...>", Short: "Create DOCX", Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error { return createDOCX(args[0], strings.Join(args[1:], " ")) },
}

var wordReadCmd = &cobra.Command{
	Use: "read <file>", Short: "Read DOCX paragraphs as text", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text, err := readDOCX(args[0])
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

var wordAppendCmd = &cobra.Command{
	Use: "append <file> <text...>", Short: "Append paragraph; writes --out or --in-place", Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateDOCX(cmd, args[0], func(document string) (string, error) {
			marker := "</w:body>"
			index := strings.LastIndex(document, marker)
			if index < 0 {
				return "", fmt.Errorf("invalid DOCX: word/document.xml has no body")
			}
			return document[:index] + docxParagraph(strings.Join(args[1:], " ")) + document[index:], nil
		})
	},
}

var wordReplaceCmd = &cobra.Command{
	Use: "replace <file> <old> <new>", Short: "Replace text; writes --out or --in-place", Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateDOCX(cmd, args[0], func(document string) (string, error) {
			old := xmlEscape(args[1])
			if !strings.Contains(document, old) {
				return "", fmt.Errorf("text not found")
			}
			return strings.ReplaceAll(document, old, xmlEscape(args[2])), nil
		})
	},
}

var wordDeleteCmd = &cobra.Command{
	Use: "delete <file> <text>", Short: "Delete matching text; writes --out or --in-place", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateDOCX(cmd, args[0], func(document string) (string, error) {
			old := xmlEscape(args[1])
			if !strings.Contains(document, old) {
				return "", fmt.Errorf("text not found")
			}
			return strings.ReplaceAll(document, old, ""), nil
		})
	},
}

func updateDOCX(cmd *cobra.Command, input string, transform func(string) (string, error)) error {
	out, _ := cmd.Flags().GetString("out")
	inPlace, _ := cmd.Flags().GetBool("in-place")
	if out == "" && !inPlace {
		return fmt.Errorf("choose --out <file> or --in-place")
	}
	if out == "" {
		out = input
	}
	return rewriteDOCX(input, out, transform)
}

func createDOCX(output, text string) error {
	contents := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
		"_rels/.rels":         `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
		"word/document.xml":   docxDocument(text),
	}
	return writeDOCX(output, contents)
}

func readDOCX(path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		input, err := file.Open()
		if err != nil {
			return "", err
		}
		defer input.Close()
		return extractDOCXText(input)
	}
	return "", fmt.Errorf("invalid DOCX: word/document.xml missing")
}

func extractDOCXText(input io.Reader) (string, error) {
	decoder := xml.NewDecoder(input)
	var builder strings.Builder
	inText := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if value.Name.Local == "t" {
				inText = false
			}
			if value.Name.Local == "p" {
				builder.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				builder.Write(value)
			}
		}
	}
	return strings.TrimRight(builder.String(), "\n"), nil
}

func rewriteDOCX(input, output string, transform func(string) (string, error)) error {
	reader, err := zip.OpenReader(input)
	if err != nil {
		return err
	}
	defer reader.Close()
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(output), ".unicli-*.docx")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	writer := zip.NewWriter(temp)
	for _, source := range reader.File {
		target, err := writer.CreateHeader(&source.FileHeader)
		if err != nil {
			writer.Close()
			temp.Close()
			return err
		}
		content, err := source.Open()
		if err != nil {
			writer.Close()
			temp.Close()
			return err
		}
		data, err := io.ReadAll(content)
		content.Close()
		if err != nil {
			writer.Close()
			temp.Close()
			return err
		}
		if source.Name == "word/document.xml" {
			updated, err := transform(string(data))
			if err != nil {
				writer.Close()
				temp.Close()
				return err
			}
			data = []byte(updated)
		}
		if _, err := target.Write(data); err != nil {
			writer.Close()
			temp.Close()
			return err
		}
	}
	if err := writer.Close(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	// zip.OpenReader keeps the input handle open on Windows. Close it before
	// replacing the input path during an explicit --in-place update.
	if err := reader.Close(); err != nil {
		return err
	}
	defer os.Remove(tempName)
	return os.Rename(tempName, output)
}

func writeDOCX(output string, files map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			return err
		}
		if _, err = io.Copy(entry, bytes.NewBufferString(content)); err != nil {
			return err
		}
	}
	return nil
}

func docxDocument(text string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + docxParagraph(text) + `<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr></w:body></w:document>`
}
func docxParagraph(text string) string {
	return `<w:p><w:r><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r></w:p>`
}
func xmlEscape(text string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(text))
	return builder.String()
}

func init() {
	for _, command := range []*cobra.Command{wordAppendCmd, wordReplaceCmd, wordDeleteCmd} {
		command.Flags().String("out", "", "output DOCX path")
		command.Flags().Bool("in-place", false, "replace input DOCX")
	}
	wordCmd.AddCommand(wordCreateCmd, wordReadCmd, wordAppendCmd, wordReplaceCmd, wordDeleteCmd)
}
