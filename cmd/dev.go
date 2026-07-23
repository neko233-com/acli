package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{Use: "dev", Short: "Cross-language build, test, format, and run templates"}
var devTestCmd = &cobra.Command{Use: "test [dir]", Short: "Run detected project test template", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runDevTemplate(projectDir(args), "test") }}
var devBuildCmd = &cobra.Command{Use: "build [dir]", Short: "Run detected project build template", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runDevTemplate(projectDir(args), "build") }}
var devFormatCmd = &cobra.Command{Use: "format [dir]", Short: "Run detected project format template", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return runDevTemplate(projectDir(args), "format") }}

func projectDir(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return "."
}
func runDevTemplate(dir, action string) error {
	var tool string
	var args []string
	if exists(filepath.Join(dir, "go.mod")) {
		tool = "go"
		switch action {
		case "test":
			args = []string{"test", "./..."}
		case "build":
			args = []string{"build", "./..."}
		case "format":
			args = []string{"fmt", "./..."}
		}
	} else if exists(filepath.Join(dir, "package.json")) {
		tool = "npm"
		switch action {
		case "test":
			args = []string{"test", "--"}
		case "build":
			args = []string{"run", "build", "--if-present"}
		case "format":
			args = []string{"run", "format", "--if-present"}
		}
	} else if exists(filepath.Join(dir, "pyproject.toml")) {
		tool = "python"
		switch action {
		case "test":
			args = []string{"-m", "pytest"}
		case "build":
			args = []string{"-m", "build"}
		case "format":
			args = []string{"-m", "ruff", "format", "."}
		}
	} else {
		return fmt.Errorf("no supported project marker in %s (go.mod, package.json, pyproject.toml)", dir)
	}
	path, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("required tool %q not found", tool)
	}
	child := exec.Command(path, args...)
	child.Dir, child.Stdin, child.Stdout, child.Stderr = dir, os.Stdin, os.Stdout, os.Stderr
	return child.Run()
}
func exists(path string) bool { _, err := os.Stat(path); return err == nil }
func init()                   { devCmd.AddCommand(devTestCmd, devBuildCmd, devFormatCmd) }
