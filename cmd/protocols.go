package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var http3TestCmd = &cobra.Command{
	Use:   "http3_test <url>",
	Short: "Test HTTP/3 support",
	Long:  `Check if a server supports HTTP/3 protocol (QUIC).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		fmt.Printf("Checking HTTP/3 support for: %s\n\n", url)

		http3Checker(url)
	},
}

var grpcTestCmd = &cobra.Command{
	Use:   "grpc_test <host> <port>",
	Short: "Test gRPC connectivity",
	Long:  `Check if a server accepts gRPC connections.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		port := args[1]
		addr := host + ":" + port

		fmt.Printf("Testing gRPC connection to: %s\n\n", addr)

		conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			return
		}
		defer conn.Close()

		fmt.Printf("gRPC endpoint reachable at %s\n", addr)
		fmt.Println("\nNote: For full gRPC testing, use grpcurl or a gRPC client.")
	},
}

var websocketTestCmd = &cobra.Command{
	Use:   "websocket_test <url>",
	Short: "Test WebSocket connection",
	Long:  `Test if a WebSocket endpoint accepts connections.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		_, _ = cmd.Flags().GetString("message")

		fmt.Printf("Testing WebSocket: %s\n\n", url)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Connection failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		if resp.StatusCode == 101 {
			fmt.Println("WebSocket upgrade successful!")
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Response: %s\n", string(body))
		}
	},
}

var socketTestCmd = &cobra.Command{
	Use:   "socket_test <host> <port>",
	Short: "Test raw socket connection",
	Long:  `Test if a TCP/UDP socket is reachable.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		port := args[1]
		protocol, _ := cmd.Flags().GetString("protocol")
		timeout := 5 * time.Second

		addr := host + ":" + port
		fmt.Printf("Testing %s socket: %s\n\n", protocol, addr)

		var err error
		var conn net.Conn

		if protocol == "udp" {
			conn, err = net.DialTimeout("udp", addr, timeout)
		} else {
			conn, err = net.DialTimeout("tcp", addr, timeout)
		}

		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			return
		}
		defer conn.Close()

		fmt.Printf("Success! Connected to %s\n", addr)
		fmt.Printf("Local address: %s\n", conn.LocalAddr().String())
	},
}

var latencyTestCmd = &cobra.Command{
	Use:   "latency_test <host>",
	Short: "Measure network latency",
	Long:  `Measure latency to a host by opening and closing connections.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		count, _ := cmd.Flags().GetInt("count")

		if count <= 0 {
			count = 10
		}

		fmt.Printf("Measuring latency to %s (%d attempts)...\n\n", host, count)

		var total time.Duration
		success := 0

		for i := 0; i < count; i++ {
			start := time.Now()
			conn, err := net.DialTimeout("tcp", host+":80", 3*time.Second)
			latency := time.Since(start)

			if err == nil {
				conn.Close()
				total += latency
				success++
				fmt.Printf("  %d: %.2fms\n", i+1, float64(latency.Microseconds())/1000)
			} else {
				fmt.Printf("  %d: timeout\n", i+1)
			}
			time.Sleep(100 * time.Millisecond)
		}

		if success > 0 {
			avg := float64(total.Microseconds()) / float64(success) / 1000
			fmt.Printf("\nAvg latency: %.2fms (%d/%d successful)\n", avg, success, count)
		}
	},
}

func http3Checker(url string) {
	fmt.Println("HTTP/3 Detection:")
	fmt.Println("  HTTP/3 uses QUIC protocol on UDP port 443")
	fmt.Println("  Full support requires quic-go library or curl with HTTP/3")
	fmt.Println("")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  HTTPS Status: %d\n", resp.StatusCode)
	fmt.Printf("  Protocol: %s\n", resp.Proto)

	if resp.TLS != nil {
		fmt.Printf("  TLS Version: %d.%d\n", resp.TLS.Version>>8, resp.TLS.Version&0xFF)
		fmt.Printf("  HTTP/3 capable server: Likely if server advertises alt-svc\n")
	}

	fmt.Println("\n  To fully test HTTP/3:")
	fmt.Println("    1. Install curl with HTTP/3 support")
	fmt.Println("    2. Or use: curl -I --http3 <url>")
}

func init() {
	websocketTestCmd.Flags().StringP("message", "m", "", "Message to send")
	socketTestCmd.Flags().StringP("protocol", "p", "tcp", "Protocol (tcp or udp)")
	latencyTestCmd.Flags().IntP("count", "c", 10, "Number of attempts")
}