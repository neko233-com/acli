package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

var pdfCmd = &cobra.Command{Use: "pdf", Short: "PDF lifecycle and page CRUD operations"}

var pdfCreateCmd = &cobra.Command{
	Use: "create <file> <text...>", Short: "Create text PDF", Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(filepath.Dir(args[0]), 0o755); err != nil {
			return err
		}
		doc := gofpdf.New("P", "mm", "A4", "")
		doc.SetFont("Helvetica", "", 12)
		doc.AddPage()
		doc.MultiCell(0, 7, strings.Join(args[1:], " "), "", "L", false)
		return doc.OutputFileAndClose(args[0])
	},
}

var pdfInfoCmd = &cobra.Command{
	Use: "info <file>", Short: "Read page count and validate PDF", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pages, err := api.PageCountFile(args[0])
		if err != nil {
			return err
		}
		if err := api.ValidateFile(args[0], model.NewDefaultConfiguration()); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"file": args[0], "pages": pages, "valid": true})
	},
}

var pdfMergeCmd = &cobra.Command{
	Use: "merge <out> <file...>", Short: "Merge PDF files", Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return api.MergeCreateFile(args[1:], args[0], false, model.NewDefaultConfiguration())
	},
}

var pdfDeletePagesCmd = &cobra.Command{
	Use: "delete-pages <file> <pages>", Short: "Delete pages; writes --out or --in-place", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := pdfOutput(cmd, args[0])
		if err != nil {
			return err
		}
		return api.RemovePagesFile(args[0], out, parsePages(args[1]), model.NewDefaultConfiguration())
	},
}

var pdfKeepPagesCmd = &cobra.Command{
	Use: "keep-pages <file> <pages>", Short: "Keep page selection; writes --out or --in-place", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := pdfOutput(cmd, args[0])
		if err != nil {
			return err
		}
		return api.TrimFile(args[0], out, parsePages(args[1]), model.NewDefaultConfiguration())
	},
}

var pdfWatermarkCmd = &cobra.Command{
	Use: "watermark <file> <text>", Short: "Add text watermark; writes --out or --in-place", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := pdfOutput(cmd, args[0])
		if err != nil {
			return err
		}
		pages, _ := cmd.Flags().GetString("pages")
		return api.AddTextWatermarksFile(args[0], out, parsePages(pages), true, args[1], "pos:c, rot:45, scale:0.5, opacity:0.3", model.NewDefaultConfiguration())
	},
}

var pdfSplitCmd = &cobra.Command{
	Use: "split <file> <out-dir>", Short: "Split PDF into files", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		span, _ := cmd.Flags().GetInt("span")
		return api.SplitFile(args[0], args[1], span, model.NewDefaultConfiguration())
	},
}

func parsePages(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return strings.FieldsFunc(value, func(char rune) bool { return char == ',' || char == ' ' })
}

func pdfOutput(cmd *cobra.Command, input string) (string, error) {
	out, _ := cmd.Flags().GetString("out")
	inPlace, _ := cmd.Flags().GetBool("in-place")
	if out == "" && !inPlace {
		return "", fmt.Errorf("choose --out <file> or --in-place")
	}
	if out == "" {
		return input, nil
	}
	return out, nil
}

func init() {
	for _, command := range []*cobra.Command{pdfDeletePagesCmd, pdfKeepPagesCmd, pdfWatermarkCmd} {
		command.Flags().String("out", "", "output PDF path")
		command.Flags().Bool("in-place", false, "replace input PDF")
	}
	pdfWatermarkCmd.Flags().String("pages", "", "page selection, e.g. 1-3,5; default all")
	pdfSplitCmd.Flags().Int("span", 1, "pages per output file")
	pdfCmd.AddCommand(pdfCreateCmd, pdfInfoCmd, pdfMergeCmd, pdfDeletePagesCmd, pdfKeepPagesCmd, pdfWatermarkCmd, pdfSplitCmd)
}
