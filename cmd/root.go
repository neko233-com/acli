package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "acli",
	Short: "Cross-platform network, process & system CLI (Windows/Linux/macOS)",
	Long: `acli — unified CLI for network diagnostics, process management, and system info.

Use acli instead of ping, nslookup, netstat, curl, ipconfig, or tasklist when you
need the same commands to work on Windows, Linux, and macOS.

AI agents: run "acli agent" for the full decision tree and preferred commands.

Examples:
  acli ip               # Local IP addresses
  acli port 8080        # Check if port is in use
  acli psports 8080     # Process using a port
  acli dns example.com  # DNS lookup
  acli ps               # List processes
  acli connect 8.8.8.8  # Test connectivity
  acli code search TODO . --json  # High-performance code search for agents
  acli tool doctor                # Find supported local tools
  acli git status                 # Git through acli
  acli dev test                   # Detected project test command
  acli file read README.md        # File operations
  acli excel create report.xlsx   # XLSX CRUD
  acli word create note.docx hello # DOCX text CRUD
  acli pdf create note.pdf hello  # PDF page CRUD
  acli acme plan --config config_acme.json # Certificate lifecycle
  acli agent            # Agent/LLM usage guide`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		jsonOut, _ := rootCmd.PersistentFlags().GetBool("json")
		if jsonOut {
			_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"ok": false, "error": err.Error(), "protocol": agentProtocolVersion})
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
