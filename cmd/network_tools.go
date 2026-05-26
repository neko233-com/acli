package cmd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var curlCmd = &cobra.Command{
	Use:   "curl <url>",
	Short: "Curl-like HTTP client",
	Long:  `Simple curl-like HTTP client with basic auth and headers support.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		method := "GET"
		if cmd.Flags().Changed("request") {
			method, _ = cmd.Flags().GetString("request")
		}
		headersStr, _ := cmd.Flags().GetString("header")
		authStr, _ := cmd.Flags().GetString("user")
		dataStr, _ := cmd.Flags().GetString("data")

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if headersStr != "" {
			for _, h := range strings.Split(headersStr, ";") {
				parts := strings.SplitN(strings.TrimSpace(h), ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
		}

		if authStr != "" {
			parts := strings.SplitN(authStr, ":", 2)
			if len(parts) == 2 {
				req.SetBasicAuth(parts[0], parts[1])
			}
		}

		if dataStr != "" {
			req.Body = io.NopCloser(strings.NewReader(dataStr))
			if req.Header.Get("Content-Type") == "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}

		fmt.Printf("%s %s\n", method, url)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("\nHTTP/1.1 %d %s\n", resp.StatusCode, resp.Status)
		for k, v := range resp.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ", "))
		}
		fmt.Printf("\n%s\n", string(body))
	},
}

var netcatCmd = &cobra.Command{
	Use:   "netcat <host> <port>",
	Short: "Netcat-like TCP client",
	Long:  `Connect to a TCP port and interact.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		port := args[1]
		addr := net.JoinHostPort(host, port)

		timeout := 10 * time.Second
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		fmt.Printf("Connected to %s\n", addr)
		fmt.Println("Type messages and press Enter to send. Press Ctrl+C to exit.\n")

		go func() {
			buf := make([]byte, 1024)
			for {
				conn.SetReadDeadline(time.Now().Add(30 * time.Second))
				n, err := conn.Read(buf)
				if err != nil {
					return
				}
				fmt.Printf("\nReceived: %s\n", string(buf[:n]))
			}
		}()

		var input string
		for {
			fmt.Print("> ")
			fmt.Scanln(&input)
			if input == "exit" {
				break
			}
			conn.Write([]byte(input + "\n"))
		}
	},
}

var portScanCmd = &cobra.Command{
	Use:   "port_scan <host> [start] [end]",
	Short: "Quick port scan",
	Long:  `Scan common ports on a host quickly.`,
	Args:  cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		start := 1
		end := 1024
		if len(args) >= 2 {
			fmt.Sscanf(args[1], "%d", &start)
		}
		if len(args) >= 3 {
			fmt.Sscanf(args[2], "%d", &end)
		}

		commonPorts := []int{21, 22, 23, 25, 53, 80, 110, 143, 443, 3306, 3389, 5432, 6379, 8080, 8443}

		fmt.Printf("Scanning %s (%d common ports)...\n", host, len(commonPorts))

		open := 0
		for _, port := range commonPorts {
			addr := fmt.Sprintf("%s:%d", host, port)
			conn, err := net.DialTimeout("tcp", addr, time.Second)
			if err == nil {
				conn.Close()
				fmt.Printf("  %3d OPEN\n", port)
				open++
			}
		}

		fmt.Printf("\n%d/%d ports open\n", open, len(commonPorts))
	},
}

func init() {
	curlCmd.Flags().StringP("request", "X", "GET", "HTTP method")
	curlCmd.Flags().StringP("header", "H", "", "Headers (key:value;key:value)")
	curlCmd.Flags().StringP("user", "u", "", "Basic auth (user:password)")
	curlCmd.Flags().StringP("data", "d", "", "Request body")
}