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

var tcpClientCmd = &cobra.Command{
	Use:   "tcp_client <host> <port>",
	Short: "TCP client",
	Long:  `Connect to a TCP server and send data.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		port := args[1]
		addr := net.JoinHostPort(host, port)
		message, _ := cmd.Flags().GetString("message")

		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			fmt.Printf("Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		fmt.Printf("Connected to %s\n", addr)

		if message != "" {
			fmt.Fprintln(conn, message)
			fmt.Printf("Sent: %s\n", message)
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		response, err := io.ReadAll(conn)
		if err == nil && len(response) > 0 {
			fmt.Printf("Received: %s\n", string(response))
		}
	},
}

var udpClientCmd = &cobra.Command{
	Use:   "udp_client <host> <port>",
	Short: "UDP client",
	Long:  `Send data to a UDP server.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		port := args[1]
		addr := net.JoinHostPort(host, port)
		message, _ := cmd.Flags().GetString("message")

		conn, err := net.DialTimeout("udp", addr, 10*time.Second)
		if err != nil {
			fmt.Printf("Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		fmt.Printf("Connected to %s\n", addr)

		if message != "" {
			fmt.Fprintln(conn, message)
			fmt.Printf("Sent: %s\n", message)
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		response := make([]byte, 1024)
		conn.Read(response)
		if len(response) > 0 {
			fmt.Printf("Received: %s\n", strings.TrimSpace(string(response)))
		}
	},
}

var httpDebugCmd = &cobra.Command{
	Use:   "http_debug <url>",
	Short: "Debug HTTP request",
	Long:  `Make HTTP request with full request/response details.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		method, _ := cmd.Flags().GetString("method")
		headersStr, _ := cmd.Flags().GetString("headers")

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if headersStr != "" {
			for _, h := range strings.Split(headersStr, ",") {
				parts := strings.SplitN(strings.TrimSpace(h), ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
		}

		fmt.Println("=== Request ===")
		fmt.Printf("Method:  %s\n", method)
		fmt.Printf("URL:     %s\n", url)
		fmt.Println("Headers:")
		for k, v := range req.Header {
			fmt.Printf("  %s: %s\n", k, v)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		fmt.Println()
		fmt.Println("=== Response ===")
		fmt.Printf("Status:  %d %s\n", resp.StatusCode, resp.Status)
		fmt.Println("Headers:")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("\nBody (%d bytes):\n%s\n", len(body), string(body))
	},
}

func init() {
	tcpClientCmd.Flags().StringP("message", "m", "", "Message to send")
	udpClientCmd.Flags().StringP("message", "m", "", "Message to send")
	httpDebugCmd.Flags().StringP("method", "X", "GET", "HTTP method")
	httpDebugCmd.Flags().StringP("headers", "H", "", "Headers (key:value,key:value)")
}
