package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns <domain>",
	Short: "DNS lookup",
	Long:  `Resolve domain name to IP addresses (A, AAAA records).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]

		fmt.Printf("Looking up DNS records for: %s\n\n", domain)

		ips, err := net.LookupIP(domain)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("IP Addresses (A/AAAA):")
		for _, ip := range ips {
			fmt.Printf("  %s (%s)\n", ip.String(), ip.To4() != nil)
		}

		cname, err := net.LookupCNAME(domain)
		if err == nil && cname != domain {
			fmt.Printf("\nCanonical Name (CNAME): %s\n", cname)
		}

		MXs, err := net.LookupMX(domain)
		if err == nil && len(MXs) > 0 {
			fmt.Println("\nMail Servers (MX):")
			for _, mx := range MXs {
				fmt.Printf("  %s (priority: %d)\n", mx.Host, mx.Pref)
			}
		}

		NSs, err := net.LookupNS(domain)
		if err == nil && len(NSs) > 0 {
			fmt.Println("\nName Servers (NS):")
			for _, ns := range NSs {
				fmt.Printf("  %s\n", ns.Host)
			}
		}
	},
}