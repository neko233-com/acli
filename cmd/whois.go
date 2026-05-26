package cmd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var whoisCmd = &cobra.Command{
	Use:   "whois <domain>",
	Short: "Whois lookup",
	Long:  `Query WHOIS database for domain registration information.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		domain = strings.TrimPrefix(domain, "https://")
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.Split(domain, "/")[0]

		fmt.Printf("Querying WHOIS for: %s\n\n", domain)

		whoisServers := []string{
			"whois.verisign.com",
			"whois.iana.org",
		}

		var whoisInfo string
		for _, server := range whoisServers {
			conn, err := net.DialTimeout("tcp", server+":43", 5*time.Second)
			if err != nil {
				continue
			}
			defer conn.Close()

			fmt.Fprintf(conn, domain+"\r\n")
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			data, err := io.ReadAll(conn)
			if err == nil {
				whoisInfo = string(data)
				break
			}
		}

		if whoisInfo == "" {
			resp, err := http.Get("https://www.whois.com/whois/" + domain)
			if err == nil {
				defer resp.Body.Close()
				data, _ := io.ReadAll(resp.Body)
				whoisInfo = string(data)
			}
		}

		if whoisInfo != "" {
			lines := strings.Split(whoisInfo, "\n")
			for _, line := range lines {
				lower := strings.ToLower(line)
				if strings.Contains(lower, "domain") ||
					strings.Contains(lower, "registrar") ||
					strings.Contains(lower, "creation") ||
					strings.Contains(lower, "expir") ||
					strings.Contains(lower, "status") ||
					strings.Contains(lower, "name server") ||
					strings.Contains(lower, "admin") ||
					strings.Contains(lower, "tech") {
					fmt.Println(line)
				}
			}
		} else {
			fmt.Println("Could not retrieve WHOIS information.")
		}
	},
}
