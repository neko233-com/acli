package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:   "lookup <host>",
	Short: "Reverse DNS lookup",
	Long:  `Perform reverse DNS lookup to find hostname from IP address.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		addr := net.ParseIP(host)
		if addr == nil {
			fmt.Printf("Invalid IP address: %s\n", host)
			return
		}

		names, err := net.LookupAddr(host)
		if err != nil {
			fmt.Printf("No PTR record found for %s\n", host)
			return
		}

		fmt.Printf("Reverse DNS for %s:\n", host)
		for _, name := range names {
			fmt.Printf("  %s\n", strings.TrimSuffix(name, "."))
		}
	},
}
