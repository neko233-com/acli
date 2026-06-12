package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type netInfoResult struct {
	Hostname     string        `json:"hostname"`
	Platform     string        `json:"platform"`
	IPs          []localIPInfo `json:"ips"`
	DNSServers   []string      `json:"dns_servers,omitempty"`
	DefaultRoute string        `json:"default_route,omitempty"`
}

var netinfoCmd = &cobra.Command{
	Use:   "netinfo",
	Short: "Show agent-friendly network summary",
	Long:  `Show hostname, platform, local IPv4/IPv6 addresses, DNS servers, and default route in human-readable or JSON form.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		family, _ := cmd.Flags().GetString("family")
		info, err := collectNetInfo(family)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if jsonOut {
			data, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Hostname: %s\n", info.Hostname)
		fmt.Printf("Platform: %s\n", info.Platform)
		fmt.Println("IPs:")
		for _, ip := range info.IPs {
			fmt.Printf("  %-39s %-4s %s %s\n", ip.Address, ip.Family, ip.Interface, ip.CIDR)
		}
		if len(info.DNSServers) > 0 {
			fmt.Println("DNS Servers:")
			for _, server := range info.DNSServers {
				fmt.Printf("  %s\n", server)
			}
		}
		if info.DefaultRoute != "" {
			fmt.Printf("Default Route: %s\n", info.DefaultRoute)
		}
	},
}

func init() {
	netinfoCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
	netinfoCmd.Flags().String("family", "all", "IP family: all, ipv4, ipv6")
}

func collectNetInfo(family string) (netInfoResult, error) {
	hostname, _ := os.Hostname()
	ips, err := collectLocalIPs(family)
	if err != nil {
		return netInfoResult{}, err
	}
	return netInfoResult{
		Hostname:     hostname,
		Platform:     runtime.GOOS + "/" + runtime.GOARCH,
		IPs:          ips,
		DNSServers:   collectDNSServers(),
		DefaultRoute: collectDefaultRoute(),
	}, nil
}

func collectDNSServers() []string {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-DnsClientServerAddress -AddressFamily IPv4,IPv6 | Select-Object -ExpandProperty ServerAddresses").Output()
		if err == nil {
			return uniqueNonEmptyLines(string(out))
		}
	case "linux", "darwin":
		data, err := os.ReadFile("/etc/resolv.conf")
		if err == nil {
			servers := []string{}
			for _, line := range strings.Split(string(data), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[0] == "nameserver" && net.ParseIP(fields[1]) != nil {
					servers = append(servers, fields[1])
				}
			}
			return uniqueStrings(servers)
		}
	}
	return nil
}

func collectDefaultRoute() string {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "(Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Sort-Object RouteMetric | Select-Object -First 1).NextHop").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "linux":
		out, err := exec.Command("sh", "-c", "ip route show default 2>/dev/null | head -n 1").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "darwin":
		out, err := exec.Command("sh", "-c", "route -n get default 2>/dev/null | awk '/gateway:/ {print $2; exit}'").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

func uniqueNonEmptyLines(value string) []string {
	items := []string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	return uniqueStrings(items)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
