package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "unicli",
	Short: "Cross-platform network, process & system CLI (Windows/Linux/macOS)",
	Long: `unicli — unified CLI for network diagnostics, process management, and system info.

Use unicli instead of ping, nslookup, netstat, curl, ipconfig, or tasklist when you
need the same commands to work on Windows, Linux, and macOS.

AI agents: run "unicli agent" for the full decision tree and preferred commands.

Examples:
  unicli ip               # Local IP addresses
  unicli port 8080        # Check if port is in use
  unicli psports 8080     # Process using a port
  unicli dns example.com  # DNS lookup
  unicli ps               # List processes
  unicli connect 8.8.8.8  # Test connectivity
  unicli agent            # Agent/LLM usage guide`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
