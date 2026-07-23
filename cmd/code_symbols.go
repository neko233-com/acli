package cmd

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type codeSymbol struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	File string `json:"file"`
	Line int    `json:"line"`
}

var codeSymbolsCmd = &cobra.Command{Use: "symbols <file|dir>", Short: "List Go AST symbols as JSON", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	files := []string{}
	info, err := os.Stat(args[0])
	if err != nil {
		return err
	}
	if info.IsDir() {
		err = filepath.WalkDir(args[0], func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && filepath.Ext(path) == ".go" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		files = append(files, args[0])
	}
	set := token.NewFileSet()
	symbols := []codeSymbol{}
	for _, path := range files {
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			switch value := decl.(type) {
			case *ast.FuncDecl:
				symbols = append(symbols, codeSymbol{Name: value.Name.Name, Kind: "func", File: path, Line: set.Position(value.Pos()).Line})
			case *ast.GenDecl:
				for _, spec := range value.Specs {
					if named, ok := spec.(*ast.TypeSpec); ok {
						symbols = append(symbols, codeSymbol{Name: named.Name.Name, Kind: "type", File: path, Line: set.Position(named.Pos()).Line})
					}
				}
			}
		}
	}
	return json.NewEncoder(os.Stdout).Encode(symbols)
}}

func init() { codeCmd.AddCommand(codeSymbolsCmd) }
