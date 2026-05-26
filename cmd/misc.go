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

var currentVersion = "v1.0.3"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates",
	Long:  `Check if a newer version of unicli is available.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", currentVersion)

		resp, err := http.Get("https://api.github.com/repos/neko233-com/unicli/releases/latest")
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
			if tag != currentVersion {
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
	Long: `Generate shell completion script for unicli.

To load completions:

Bash:

  $ source <(unicli completion bash)

  # To load completions for each session, execute once:

  # Linux:
  $ unicli completion bash > /etc/bash_completion.d/unicli

  # macOS:
  $ unicli completion bash > /usr/local/etc/bash_completion.d/unicli

Zsh:

  # If shell completion is not enabled in your environment,
  # you will need to add the following line to your .zshrc:

  # autoload - compinit; compinit

  $ unicli completion zsh > "${fpath[1]}/_unicli"

Fish:

  $ unicli completion fish | source

  # To load completions for each session, execute once:
  $ unicli completion fish > ~/.config/fish/completions/unicli.fish`,
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
	Long:  `Display version and build information for unicli.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("unicli %s\n", currentVersion)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("\nWebsite: https://github.com/neko233-com/unicli")
	},
}

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Open documentation",
	Long:  `Open the unicli documentation in browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		docURL := "https://github.com/neko233-com/unicli#readme"
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
