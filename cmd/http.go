package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http <url>",
	Short: "Make HTTP request",
	Long:  `Make an HTTP/HTTPS request and show response details including headers, status, and timing.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		urlStr := args[0]
		method, _ := cmd.Flags().GetString("method")
		followRedirects, _ := cmd.Flags().GetBool("follow")
		showHeaders, _ := cmd.Flags().GetBool("headers")

		if !strings.HasPrefix(urlStr, "http") {
			urlStr = "https://" + urlStr
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			fmt.Printf("Invalid URL: %v\n", err)
			os.Exit(1)
		}

		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		if !followRedirects {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}

		req, err := http.NewRequest(method, parsedURL.String(), nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}

		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Request failed: %v\n", err)
			os.Exit(1)
		}
		duration := time.Since(start)

		fmt.Printf("Status:  %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("Time:    %v\n", duration)
		fmt.Printf("Size:    %d bytes\n", resp.ContentLength)
		fmt.Println()

		if showHeaders {
			fmt.Println("Headers:")
			for key, values := range resp.Header {
				fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
			}
			fmt.Println()
		}

		if parsedURL.Scheme == "https" {
			fmt.Println("TLS Info:")
			fmt.Printf("  Protocol: %s\n", resp.TLS.Version)
			fmt.Printf("  Cipher:   %s\n", resp.TLS.CipherSuite)
			fmt.Printf("  Server:   %s\n", resp.TLS.ServerName)
		}
	},
}

func init() {
	httpCmd.Flags().StringP("method", "X", "GET", "HTTP method")
	httpCmd.Flags().Bool("follow", false, "Follow redirects")
	httpCmd.Flags().Bool("headers", false, "Show response headers")
}

var httpsCheckCmd = &cobra.Command{
	Use:   "ssl <host>",
	Short: "Check SSL certificate",
	Long:  `Check SSL certificate details for an HTTPS host.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		port := 443

		if strings.Contains(host, ":") {
			parts := strings.Split(host, ":")
			host = parts[0]
			fmt.Sscanf(parts[1], "%d", &port)
		}

		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), nil)
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		cert := conn.ConnectionState().PeerCertificates[0]
		fmt.Printf("SSL Certificate for %s:%d\n\n", host, port)
		fmt.Printf("Subject:    %s\n", cert.Subject.CommonName)
		fmt.Printf("Issuer:     %s\n", cert.Issuer.CommonName)
		fmt.Printf("Valid From: %s\n", cert.NotBefore.Format("2006-01-02 15:04:05"))
		fmt.Printf("Valid To:   %s\n", cert.NotAfter.Format("2006-01-02 15:04:05"))
		fmt.Printf("Serial:     %x\n", cert.SerialNumber)
	},
}