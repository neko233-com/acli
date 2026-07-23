package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCodeSymbolsHandlerAndFileCopy(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(source, []byte("package sample\n\ntype Item struct{}\nfunc Run() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	codeSymbolsCmd.SetOut(&output)
	defer codeSymbolsCmd.SetOut(nil)
	if err := codeSymbolsCmd.RunE(codeSymbolsCmd, []string{source}); err != nil {
		t.Fatal(err)
	}
	var symbols []codeSymbol
	if err := json.Unmarshal(output.Bytes(), &symbols); err != nil || len(symbols) != 2 || symbols[0].Name != "Item" || symbols[1].Name != "Run" {
		t.Fatalf("symbols=%s err=%v", output.String(), err)
	}
	destination := filepath.Join(dir, "nested", "copy.go")
	if err := copyLocalFile(source, destination); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "package sample\n\ntype Item struct{}\nfunc Run() {}\n" {
		t.Fatalf("copy=%q err=%v", data, err)
	}
	if err := copyLocalFile(dir, filepath.Join(dir, "bad")); err == nil {
		t.Fatal("expected directory copy rejection")
	}
}

func TestFileWriteAppendAndPatchHandlers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "note.txt")
	if err := fileWriteCmd.RunE(fileWriteCmd, []string{path, "hello"}); err != nil {
		t.Fatal(err)
	}
	if err := fileWriteCmd.RunE(fileWriteCmd, []string{path, "again"}); err == nil {
		t.Fatal("expected overwrite protection")
	}
	if err := fileAppendCmd.RunE(fileAppendCmd, []string{path, "world"}); err != nil {
		t.Fatal(err)
	}
	patchedPath := filepath.Join(dir, "patched.txt")
	rootCmd.SetArgs([]string{"file", "patch", path, "world", "acli", "--out", patchedPath})
	defer func() {
		rootCmd.SetArgs(nil)
		_ = rootCmd.PersistentFlags().Set("out", "")
	}()
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	patched, err := os.ReadFile(patchedPath)
	if err != nil || string(patched) != "hello\nacli\n" {
		t.Fatalf("patched=%q err=%v", patched, err)
	}
}
