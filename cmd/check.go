package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type checkResult struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	OK      bool   `json:"ok"`
	Detail  string `json:"detail,omitempty"`
	Error   string `json:"error,omitempty"`
	Latency string `json:"latency,omitempty"`
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run declarative ops checks",
	Long:  `Run port, DNS, HTTP, and disk threshold checks. Results are JSON by default for agents.`,
	Run: func(cmd *cobra.Command, args []string) {
		ports, _ := cmd.Flags().GetStringSlice("port")
		domains, _ := cmd.Flags().GetStringSlice("dns")
		urls, _ := cmd.Flags().GetStringSlice("http")
		diskMax, _ := cmd.Flags().GetFloat64("disk-max")
		jsonOut, _ := cmd.Flags().GetBool("json")
		results := []checkResult{}
		for _, item := range ports {
			results = append(results, runPortCheck(item))
		}
		for _, domain := range domains {
			results = append(results, runDNSCheck(domain))
		}
		for _, url := range urls {
			results = append(results, runHTTPCheck(url))
		}
		if diskMax > 0 {
			results = append(results, runDiskCheck(diskMax)...)
		}
		if jsonOut {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
			return
		}
		for _, result := range results {
			fmt.Printf("%-5v %-6s %-30s %s %s\n", result.OK, result.Type, result.Name, result.Detail, result.Error)
		}
		for _, result := range results {
			if !result.OK {
				os.Exit(1)
			}
		}
	},
}

func init() {
	checkCmd.Flags().StringSlice("port", nil, "TCP check host:port (repeatable)")
	checkCmd.Flags().StringSlice("dns", nil, "DNS check domain (repeatable)")
	checkCmd.Flags().StringSlice("http", nil, "HTTP check URL (repeatable)")
	checkCmd.Flags().Float64("disk-max", 0, "Fail if any disk used percent exceeds this value")
	checkCmd.Flags().Bool("json", true, "Output JSON")
}

func runPortCheck(item string) checkResult {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", item, 3*time.Second)
	latency := time.Since(start).String()
	if err != nil {
		return checkResult{Name: item, Type: "port", OK: false, Error: err.Error(), Latency: latency}
	}
	conn.Close()
	return checkResult{Name: item, Type: "port", OK: true, Detail: "open", Latency: latency}
}

func runDNSCheck(domain string) checkResult {
	start := time.Now()
	ips, err := net.LookupIP(domain)
	latency := time.Since(start).String()
	if err != nil {
		return checkResult{Name: domain, Type: "dns", OK: false, Error: err.Error(), Latency: latency}
	}
	a := 0
	aaaa := 0
	for _, ip := range ips {
		if ip.To4() != nil {
			a++
		} else {
			aaaa++
		}
	}
	return checkResult{Name: domain, Type: "dns", OK: len(ips) > 0, Detail: fmt.Sprintf("a=%d aaaa=%d", a, aaaa), Latency: latency}
}

func runHTTPCheck(url string) checkResult {
	start := time.Now()
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	latency := time.Since(start).String()
	if err != nil {
		return checkResult{Name: url, Type: "http", OK: false, Error: err.Error(), Latency: latency}
	}
	defer resp.Body.Close()
	ok := resp.StatusCode >= 200 && resp.StatusCode < 400
	return checkResult{Name: url, Type: "http", OK: ok, Detail: resp.Status, Latency: latency}
}

func runDiskCheck(maxUsed float64) []checkResult {
	results := []checkResult{}
	for _, disk := range collectDiskHealth() {
		usedText := strings.TrimSuffix(disk.UsedPercent, "%")
		used, err := strconv.ParseFloat(usedText, 64)
		if err != nil {
			continue
		}
		results = append(results, checkResult{
			Name:   disk.Name,
			Type:   "disk",
			OK:     used <= maxUsed,
			Detail: fmt.Sprintf("used=%.1f%% max=%.1f%%", used, maxUsed),
		})
	}
	return results
}
