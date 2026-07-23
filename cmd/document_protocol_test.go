package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

func TestExcelWorkbookRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.xlsx")
	book := excelize.NewFile()
	if err := book.SetCellValue("Sheet1", "A1", "acli"); err != nil {
		t.Fatal(err)
	}
	if err := saveWorkbook(book, path); err != nil {
		t.Fatal(err)
	}
	book.Close()
	reopened, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	value, err := reopened.GetCellValue("Sheet1", "A1")
	if err != nil || value != "acli" {
		t.Fatalf("value=%q err=%v", value, err)
	}
	rows, err := readExcelRange(reopened, "Sheet1", "A1:B2")
	if err != nil || rows[0][0] != "acli" {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}
func TestPDFCreateAndPageCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.pdf")
	doc := gofpdf.New("P", "mm", "A4", "")
	doc.AddPage()
	doc.SetFont("Helvetica", "", 12)
	doc.Cell(10, 10, "acli")
	if err := doc.OutputFileAndClose(path); err != nil {
		t.Fatal(err)
	}
	count, err := api.PageCountFile(path)
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}
func TestProtocolSchemaAndCommandLookup(t *testing.T) {
	schema := commandSchema(filePatchCmd)
	if schema["protocol"] != agentProtocolVersion || schema["command"] != "acli file patch" {
		t.Fatalf("bad schema: %#v", schema)
	}
	for _, path := range [][]string{{"download", "get"}, {"atlas", "pack"}, {"web", "snapshot"}, {"ffmpeg", "convert"}, {"code", "symbols"}, {"git", "push"}} {
		command, remaining, err := rootCmd.Find(path)
		if err != nil || len(remaining) != 0 || command == rootCmd {
			t.Fatalf("missing command %v: %v %v", path, remaining, err)
		}
	}
}
func TestHelpers(t *testing.T) {
	if got := parsePages("1-3, 5"); len(got) != 2 || got[0] != "1-3" {
		t.Fatalf("pages=%v", got)
	}
	profiles, command := splitProfilesAndCommand([]string{"one", "two", "--", "uname", "-a"})
	if len(profiles) != 2 || len(command) != 2 {
		t.Fatalf("split=%v %v", profiles, command)
	}
	path := filepath.Join(t.TempDir(), "hash.txt")
	if err := os.WriteFile(path, []byte("acli"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkSHA256(path, "f4c0a50c838f5f79084e87f7c3b42d41191a7ff946ce08b820434d3d8f9e3da6"); err == nil {
		t.Fatal("expected known-invalid checksum")
	}
}

func TestDocumentMutationTargetsAndValidation(t *testing.T) {
	command := &cobra.Command{}
	command.Flags().String("out", "", "")
	command.Flags().Bool("in-place", false, "")
	if _, err := pdfOutput(command, "input.pdf"); err == nil {
		t.Fatal("expected explicit PDF target requirement")
	}
	if err := command.Flags().Set("out", "output.pdf"); err != nil {
		t.Fatal(err)
	}
	if target, err := pdfOutput(command, "input.pdf"); err != nil || target != "output.pdf" {
		t.Fatalf("target=%q err=%v", target, err)
	}
	if err := command.Flags().Set("out", ""); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("in-place", "true"); err != nil {
		t.Fatal(err)
	}
	if target, err := pdfOutput(command, "input.pdf"); err != nil || target != "input.pdf" {
		t.Fatalf("in-place target=%q err=%v", target, err)
	}

	book := excelize.NewFile()
	defer book.Close()
	if _, err := readExcelRange(book, "Sheet1", "B2:A1"); err == nil {
		t.Fatal("expected reversed range failure")
	}
	if _, err := readExcelRange(book, "Sheet1", "A1:B2:C3"); err == nil {
		t.Fatal("expected malformed range failure")
	}
}

func TestDocumentCommandsUseExplicitOutput(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "source.docx")
	out := filepath.Join(dir, "result.docx")
	if err := createDOCX(doc, "hello"); err != nil {
		t.Fatal(err)
	}
	if err := wordAppendCmd.Flags().Set("out", out); err != nil {
		t.Fatal(err)
	}
	defer wordAppendCmd.Flags().Set("out", "")
	if err := wordAppendCmd.RunE(wordAppendCmd, []string{doc, "world"}); err != nil {
		t.Fatal(err)
	}
	text, err := readDOCX(out)
	if err != nil || text != "hello\nworld" {
		t.Fatalf("text=%q err=%v", text, err)
	}
	if err := wordReplaceCmd.Flags().Set("out", filepath.Join(dir, "replaced.docx")); err != nil {
		t.Fatal(err)
	}
	defer wordReplaceCmd.Flags().Set("out", "")
	if err := wordReplaceCmd.RunE(wordReplaceCmd, []string{out, "missing", "x"}); err == nil {
		t.Fatal("expected missing Word text failure")
	}
}
