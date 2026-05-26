package cmd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var ifaceCmd = &cobra.Command{
	Use:   "iface",
	Short: "Show network interfaces",
	Long:  `Display detailed information about all network interfaces including IP addresses, MAC addresses, and status.`,
	Run: func(cmd *cobra.Command, args []string) {
		ifaces, err := net.Interfaces()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		for _, iface := range ifaces {
			fmt.Printf("Interface: %s\n", iface.Name)
			fmt.Printf("  Flags:   %v\n", iface.Flags)

			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					fmt.Printf("  Addr:    %s\n", addr.String())
				}
			}

			fmt.Printf("  MTU:     %d\n", iface.MTU)

			if runtime.GOOS != "windows" {
				mac := iface.HardwareAddr.String()
				if mac != "" {
					fmt.Printf("  MAC:     %s\n", mac)
				}
			}
			fmt.Println()
		}
	},
}

var publicipCmd = &cobra.Command{
	Use:   "publicip",
	Short: "Show public IP address",
	Long:  `Query external services to determine the machine's public IP address.`,
	Run: func(cmd *cobra.Command, args []string) {
		services := []string{
			"https://api.ipify.org",
			"https://icanhazip.com",
			"https://ifconfig.me/ip",
		}

		for _, service := range services {
			resp, err := http.Get(service)
			if err == nil {
				defer resp.Body.Close()
				data, err := io.ReadAll(resp.Body)
				if err == nil {
					ip := strings.TrimSpace(string(data))
					if net.ParseIP(ip) != nil {
						fmt.Printf("Public IP: %s (via %s)\n", ip, service)
						return
					}
				}
			}
		}
		fmt.Println("Could not determine public IP")
	},
}
