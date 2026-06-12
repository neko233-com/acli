package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

type dnsLookupResult struct {
	Domain string      `json:"domain"`
	A      []string    `json:"a"`
	AAAA   []string    `json:"aaaa"`
	CNAME  string      `json:"cname,omitempty"`
	MX     []dnsMXInfo `json:"mx,omitempty"`
	NS     []string    `json:"ns,omitempty"`
}

type dnsMXInfo struct {
	Host     string `json:"host"`
	Priority uint16 `json:"priority"`
}

var dnsCmd = &cobra.Command{
	Use:   "dns <domain>",
	Short: "DNS lookup",
	Long:  `Resolve domain name to IP addresses (A, AAAA records).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		family, _ := cmd.Flags().GetString("family")
		jsonOut, _ := cmd.Flags().GetBool("json")

		result, err := lookupDNS(domain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if family == "ipv4" {
			result.AAAA = nil
		}
		if family == "ipv6" {
			result.A = nil
		}
		if jsonOut {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("DNS: %s\n", result.Domain)
		if len(result.A) > 0 {
			fmt.Println("A:")
			for _, ip := range result.A {
				fmt.Printf("  %s\n", ip)
			}
		}
		if len(result.AAAA) > 0 {
			fmt.Println("AAAA:")
			for _, ip := range result.AAAA {
				fmt.Printf("  %s\n", ip)
			}
		}
		if result.CNAME != "" {
			fmt.Printf("CNAME: %s\n", result.CNAME)
		}
		if len(result.MX) > 0 {
			fmt.Println("MX:")
			for _, mx := range result.MX {
				fmt.Printf("  %s priority=%d\n", mx.Host, mx.Priority)
			}
		}
		if len(result.NS) > 0 {
			fmt.Println("NS:")
			for _, ns := range result.NS {
				fmt.Printf("  %s\n", ns)
			}
		}
	},
}

func init() {
	dnsCmd.Flags().String("family", "all", "IP family: all, ipv4, ipv6")
	dnsCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
}

func lookupDNS(domain string) (dnsLookupResult, error) {
	result := dnsLookupResult{Domain: domain, A: []string{}, AAAA: []string{}}
	ips, err := net.LookupIP(domain)
	if err != nil {
		return result, err
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			result.A = append(result.A, ip.String())
		} else {
			result.AAAA = append(result.AAAA, ip.String())
		}
	}
	if cname, err := net.LookupCNAME(domain); err == nil && cname != domain {
		result.CNAME = cname
	}
	if mxs, err := net.LookupMX(domain); err == nil {
		for _, mx := range mxs {
			result.MX = append(result.MX, dnsMXInfo{Host: mx.Host, Priority: mx.Pref})
		}
	}
	if nss, err := net.LookupNS(domain); err == nil {
		for _, ns := range nss {
			result.NS = append(result.NS, ns.Host)
		}
	}
	return result, nil
}
