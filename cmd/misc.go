package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates",
	Long:  `Check if a newer version of netgo is available.`,
	Run: func(cmd *cobra.Command, args []string) {
		currentVersion := "v!NEW_VERSION!"
		fmt.Printf("Current version: %s\n", currentVersion)

		resp, err := http.Get("https://api.github.com/repos/neko233-com/netgo/releases/latest")
		if err != nil {
			fmt.Printf("Could not check for updates: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		data, _ := io.ReadAll(resp.Body)
		json.Unmarshal(data, &result)

		if tag, ok := result["tag_name"].(string); ok {
			fmt.Printf("Latest version: %s\n", tag)
			if tag != "v"+currentVersion {
				fmt.Println("A newer version is available!")
			} else {
				fmt.Println("You are running the latest version.")
			}
		}
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate shell completion script for netgo.

To load completions:

Bash:

  $ source <(netgo completion bash)

  # To load completions for each session, execute once:

  # Linux:
  $ netgo completion bash > /etc/bash_completion.d/netgo

  # macOS:
  $ netgo completion bash > /usr/local/etc/bash_completion.d/netgo

Zsh:

  # If shell completion is not enabled in your environment,
  # you will need to add the following line to your .zshrc:

  # autoload - compinit; compinit

  $ netgo completion zsh > "${fpath[1]}/_netgo"

Fish:

  $ netgo completion fish | source

  # To load completions for each session, execute once:
  $ netgo completion fish > ~/.config/fish/completions/netgo.fish`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		switch shell {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			fmt.Printf("Unsupported shell: %s\n", shell)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version and build information for netgo.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("netgo v1.0.0")
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("\nWebsite: https://github.com/neko233-com/netgo")
	},
}

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Open documentation",
	Long:  `Open the netgo documentation in browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		docURL := "https://github.com/neko233-com/netgo#readme"
		fmt.Printf("Opening: %s\n", docURL)

		var cmdExec *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("cmd", "/c", "start", docURL)
		case "linux":
			cmdExec = exec.Command("xdg-open", docURL)
		case "darwin":
			cmdExec = exec.Command("open", docURL)
		}
		cmdExec.Start()
	},
}