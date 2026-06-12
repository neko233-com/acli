package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

type localIPInfo struct {
	Interface string `json:"interface"`
	Address   string `json:"address"`
	Family    string `json:"family"`
	CIDR      string `json:"cidr"`
}

var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Show IP addresses",
	Long:  `Display all local IP addresses and optionally the public IP.`,
	Run: func(cmd *cobra.Command, args []string) {
		family, _ := cmd.Flags().GetString("family")
		jsonOut, _ := cmd.Flags().GetBool("json")

		ips, err := collectLocalIPs(family)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if jsonOut {
			data, _ := json.MarshalIndent(ips, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Println("Local IPs:")
		for _, ip := range ips {
			fmt.Printf("  %-39s %-4s %s %s\n", ip.Address, ip.Family, ip.Interface, ip.CIDR)
		}
	},
}

func init() {
	ipCmd.Flags().String("family", "all", "IP family: all, ipv4, ipv6")
	ipCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
}

func collectLocalIPs(family string) ([]localIPInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	result := []localIPInfo{}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ipFamily := "ipv6"
			if ipnet.IP.To4() != nil {
				ipFamily = "ipv4"
			}
			if family != "all" && family != ipFamily {
				continue
			}
			result = append(result, localIPInfo{
				Interface: iface.Name,
				Address:   ipnet.IP.String(),
				Family:    ipFamily,
				CIDR:      ipnet.String(),
			})
		}
	}
	return result, nil
}
