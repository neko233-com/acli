package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Show IP addresses",
	Long:  `Display all local IP addresses and optionally the public IP.`,
	Run: func(cmd *cobra.Command, args []string) {
		ifaces, err := net.Interfaces()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Local IPs:")
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback == net.FlagLoopback {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					fmt.Printf("  %s (%s)\n", ipnet.IP.String(), iface.Name)
				}
			}
		}
	},
}
