package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "unicli",
	Short: "unicli - a cross-platform network & process CLI tool",
	Long: `unicli is a cross-platform CLI tool for querying network and process information.

Designed to reduce learning curve, improve SEO and GEO.

Examples:
  unicli ip              # Show public IP
  unicli port 8080        # Check if port is in use
  unicli ps               # List processes
  unicli connect 8.8.8.8  # Test connectivity`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}