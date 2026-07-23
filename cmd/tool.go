package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type toolStatus struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
}

var managedTools = []string{"rg", "git", "go", "node", "npm", "pnpm", "bun", "python", "python3", "pip", "cargo", "rustc", "java"}
var safeToolName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

var toolCmd = &cobra.Command{Use: "tool", Short: "Run installed developer tools safely, without a shell"}

var toolDoctorCmd = &cobra.Command{
	Use: "doctor", Short: "Detect supported local tools and versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		items := make([]toolStatus, 0, len(managedTools))
		for _, name := range managedTools {
			items = append(items, detectTool(name))
		}
		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"platform": runtime.GOOS + "/" + runtime.GOARCH, "tools": items})
		}
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		for _, item := range items {
			state := "missing"
			if item.Available {
				state = item.Path
				if item.Version != "" {
					state += " (" + item.Version + ")"
				}
			}
			fmt.Printf("%-10s %s\n", item.Name, state)
		}
		return nil
	},
}

var toolRunCmd = &cobra.Command{
	Use: "run <tool> -- [arguments...]", Short: "Run installed tool directly; shell expansion is disabled", Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !safeToolName.MatchString(args[0]) {
			return fmt.Errorf("invalid tool name %q", args[0])
		}
		if args[0] == "sh" || args[0] == "bash" || args[0] == "zsh" || args[0] == "cmd" || args[0] == "powershell" || args[0] == "pwsh" {
			return fmt.Errorf("shell tools are blocked; pass an executable directly")
		}
		path, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("tool %q not found in PATH", args[0])
		}
		child := exec.Command(path, args[1:]...)
		child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
		return child.Run()
	},
}

func detectTool(name string) toolStatus {
	item := toolStatus{Name: name}
	path, err := exec.LookPath(name)
	if err != nil {
		return item
	}
	item.Available, item.Path = true, path
	args := []string{"--version"}
	if name == "go" {
		args = []string{"version"}
	}
	if name == "java" {
		args = []string{"-version"}
	}
	version := exec.Command(path, args...)
	output, err := version.CombinedOutput()
	if err == nil {
		item.Version = strings.TrimSpace(strings.Split(string(output), "\n")[0])
	}
	return item
}

func init() {
	sort.Strings(managedTools)
	toolDoctorCmd.Flags().Bool("json", false, "emit machine-readable tool inventory")
	toolCmd.AddCommand(toolDoctorCmd, toolRunCmd)
}
