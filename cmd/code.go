package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// codeCmd exposes ripgrep without a shell. This preserves rg's performance and
// makes every argument safe for agents to pass through programmatically.
var codeCmd = &cobra.Command{
	Use:   "code",
	Short: "High-performance code operations (ripgrep-backed)",
}

var codeSearchCmd = &cobra.Command{
	Use:   "search <pattern> [path]",
	Short: "Search code with ripgrep; raw rg output by default",
	Long: `Search files using ripgrep (rg), the fastest supported search backend.

Use --json for agent parsing. Arguments are passed directly to rg, never through
a shell. Install ripgrep to use this command.

Examples:
  acli code search 'TODO|FIXME' . --glob '*.go'
  acli code search 'func main' . --json
  acli code search password . --hidden --ignore-case`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		rg, err := exec.LookPath("rg")
		if err != nil {
			return fmt.Errorf("ripgrep (rg) not found in PATH; install it for acli code search")
		}

		path := "."
		if len(args) == 2 {
			path = args[1]
		}
		rgArgs := []string{"--line-number", "--column", "--color=never", "--no-heading"}
		if cmd.Flags().Changed("json") {
			rgArgs = append(rgArgs, "--json")
		}
		if cmd.Flags().Changed("ignore-case") {
			rgArgs = append(rgArgs, "--ignore-case")
		}
		if cmd.Flags().Changed("fixed-strings") {
			rgArgs = append(rgArgs, "--fixed-strings")
		}
		if cmd.Flags().Changed("hidden") {
			rgArgs = append(rgArgs, "--hidden")
		}
		for _, glob := range mustStringSlice(cmd, "glob") {
			rgArgs = append(rgArgs, "--glob", glob)
		}
		if max, _ := cmd.Flags().GetInt("max-count"); max > 0 {
			rgArgs = append(rgArgs, "--max-count", fmt.Sprint(max))
		}
		rgArgs = append(rgArgs, "--", args[0], path)

		search := exec.Command(rg, rgArgs...)
		search.Stdin, search.Stdout, search.Stderr = os.Stdin, os.Stdout, os.Stderr
		err = search.Run()
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
			return nil // rg convention: no matches.
		}
		return err
	},
}

var codeFilesCmd = &cobra.Command{
	Use:   "files [path]",
	Short: "List files tracked by ripgrep",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rg, err := exec.LookPath("rg")
		if err != nil {
			return fmt.Errorf("ripgrep (rg) not found in PATH; install it for acli code files")
		}
		path := "."
		if len(args) == 1 {
			path = args[0]
		}
		files := exec.Command(rg, "--files", path)
		files.Stdout, files.Stderr = os.Stdout, os.Stderr
		return files.Run()
	},
}

func mustStringSlice(cmd *cobra.Command, name string) []string {
	values, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func init() {
	codeSearchCmd.Flags().Bool("json", false, "emit ripgrep JSON Lines for agents")
	codeSearchCmd.Flags().BoolP("ignore-case", "i", false, "case-insensitive search")
	codeSearchCmd.Flags().BoolP("fixed-strings", "F", false, "treat pattern as literal text")
	codeSearchCmd.Flags().Bool("hidden", false, "search hidden files (still respects ignore files)")
	codeSearchCmd.Flags().StringSliceP("glob", "g", nil, "include or exclude glob; repeatable")
	codeSearchCmd.Flags().IntP("max-count", "m", 0, "max matches per file")
	codeCmd.AddCommand(codeSearchCmd, codeFilesCmd)
}
