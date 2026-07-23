package cmd

import (
	"archive/zip"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAtlasPackSplitRoundTrip(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "src")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"hero.png", "enemy.png"} {
		imageFile, err := os.Create(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		sprite := image.NewRGBA(image.Rect(0, 0, 5, 7))
		sprite.Set(0, 0, color.RGBA{R: 255, A: 255})
		if err := png.Encode(imageFile, sprite); err != nil {
			t.Fatal(err)
		}
		imageFile.Close()
	}
	atlas, manifest := filepath.Join(dir, "atlas.png"), filepath.Join(dir, "atlas.json")
	if err := packAtlas(source, atlas, manifest, 1, 32); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "out")
	if err := splitAtlas(atlas, manifest, output); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"hero.png", "enemy.png"} {
		file, err := os.Open(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := png.Decode(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		if decoded.Bounds().Dx() != 5 || decoded.Bounds().Dy() != 7 {
			t.Fatalf("unexpected split size: %v", decoded.Bounds())
		}
	}
}

func TestDOCXCreateReadAndRewrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.docx")
	if err := createDOCX(path, "hello"); err != nil {
		t.Fatal(err)
	}
	if err := rewriteDOCX(path, path, func(document string) (string, error) { return strings.Replace(document, "hello", "world", 1), nil }); err != nil {
		t.Fatal(err)
	}
	got, err := readDOCX(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "world" {
		t.Fatalf("got %q", got)
	}
}

func TestZipRoundTripAndSlipProtection(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input")
	if err := os.Mkdir(input, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(input, "a.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "bundle.zip")
	if err := createZip(archive, []string{input}); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "out")
	if err := extractZip(archive, output); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(output, "input", "a.txt")); err != nil {
		t.Fatal(err)
	}
	slip := filepath.Join(dir, "slip.zip")
	file, err := os.Create(slip)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, _ := writer.Create("../escape.txt")
	_, _ = entry.Write([]byte("no"))
	writer.Close()
	file.Close()
	if err := extractZip(slip, filepath.Join(dir, "safe")); err == nil {
		t.Fatal("expected zip slip rejection")
	}
}

func TestRemotePathParsing(t *testing.T) {
	if parseRemotePath(`C:\tmp\file`) != nil {
		t.Fatal("windows path treated as remote")
	}
	parsed := parseRemotePath("user@example.com:/tmp/file")
	if parsed == nil || parsed.user != "user" || parsed.host != "example.com" || parsed.path != "/tmp/file" {
		t.Fatalf("bad remote path: %#v", parsed)
	}
	if parseRemotePath("local.txt") != nil {
		t.Fatal("local file treated as remote")
	}
}

func TestCommandTreeHasUsableMetadata(t *testing.T) {
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		if command.Name() == "" || command.Short == "" {
			t.Fatalf("bad command %q", command.CommandPath())
		}
		for _, child := range command.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}
