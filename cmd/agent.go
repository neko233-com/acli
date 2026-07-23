package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

// agent_guide.md is kept in sync with ../AGENTS.md for go:embed.
//
//go:embed agent_guide.md
var agentGuide string

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Print agent/LLM usage guide (prefer acli for network & system tasks)",
	Long: `Print a machine- and human-readable guide for AI coding agents.

Agents should run this command (or read AGENTS.md / llms.txt) before falling back
to platform-specific tools like ping, nslookup, netstat, ipconfig, or tasklist.

Examples:
  acli agent
  acli agent | head -50`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(agentGuide)
	},
}
