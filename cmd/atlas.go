package cmd

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type atlasFrame struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}
type atlasManifest struct {
	Width   int                   `json:"width"`
	Height  int                   `json:"height"`
	Padding int                   `json:"padding"`
	Frames  map[string]atlasFrame `json:"frames"`
}

var atlasCmd = &cobra.Command{Use: "atlas", Short: "Offline game sprite atlas pack and split"}
var atlasPackCmd = &cobra.Command{Use: "pack <input-dir> <out.png> <out.json>", Short: "Pack PNG/JPEG sprites into deterministic atlas", Args: cobra.ExactArgs(3), RunE: func(cmd *cobra.Command, args []string) error {
	padding, _ := cmd.Flags().GetInt("padding")
	max, _ := cmd.Flags().GetInt("max-size")
	return packAtlas(args[0], args[1], args[2], padding, max)
}}
var atlasSplitCmd = &cobra.Command{Use: "split <atlas.png> <manifest.json> <out-dir>", Short: "Split atlas using its JSON manifest", Args: cobra.ExactArgs(3), RunE: func(cmd *cobra.Command, args []string) error { return splitAtlas(args[0], args[1], args[2]) }}

func packAtlas(input, outPNG, outJSON string, padding, max int) error {
	entries, err := os.ReadDir(input)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	canvas := image.NewRGBA(image.Rect(0, 0, max, max))
	manifest := atlasManifest{Width: max, Height: max, Padding: padding, Frames: map[string]atlasFrame{}}
	x, y, row := padding, padding, 0
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".png" && filepath.Ext(entry.Name()) != ".jpg" && filepath.Ext(entry.Name()) != ".jpeg") {
			continue
		}
		file, err := os.Open(filepath.Join(input, entry.Name()))
		if err != nil {
			return err
		}
		sprite, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			return err
		}
		b := sprite.Bounds()
		w, h := b.Dx(), b.Dy()
		if w+padding*2 > max || h+padding*2 > max {
			return fmt.Errorf("sprite %s exceeds max-size", entry.Name())
		}
		if x+w+padding > max {
			x = padding
			y += row + padding
			row = 0
		}
		if y+h+padding > max {
			return fmt.Errorf("atlas exceeds max-size %d", max)
		}
		draw.Draw(canvas, image.Rect(x, y, x+w, y+h), sprite, b.Min, draw.Over)
		manifest.Frames[strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))] = atlasFrame{x, y, w, h}
		x += w + padding
		if h > row {
			row = h
		}
	}
	file, err := os.Create(outPNG)
	if err != nil {
		return err
	}
	err = png.Encode(file, canvas)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outJSON, data, 0o644)
}
func splitAtlas(pngPath, jsonPath, outDir string) error {
	file, err := os.Open(pngPath)
	if err != nil {
		return err
	}
	sheet, err := png.Decode(file)
	file.Close()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var manifest atlasManifest
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if err = os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for name, frame := range manifest.Frames {
		out := image.NewRGBA(image.Rect(0, 0, frame.W, frame.H))
		draw.Draw(out, out.Bounds(), sheet, image.Point{X: frame.X, Y: frame.Y}, draw.Src)
		target, err := os.Create(filepath.Join(outDir, name+".png"))
		if err != nil {
			return err
		}
		err = png.Encode(target, out)
		closeErr := target.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
func init() {
	atlasPackCmd.Flags().Int("padding", 2, "transparent padding in pixels")
	atlasPackCmd.Flags().Int("max-size", 2048, "square atlas maximum size")
	atlasCmd.AddCommand(atlasPackCmd, atlasSplitCmd)
}
