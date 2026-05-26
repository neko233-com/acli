package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var httpPostCmd = &cobra.Command{
	Use:   "http_post <url>",
	Short: "POST data to HTTP endpoint",
	Long:  `Send POST request with data to an HTTP endpoint.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		data, _ := cmd.Flags().GetString("data")
		contentType, _ := cmd.Flags().GetString("content-type")

		if data == "" {
			data = "{}"
		}
		if contentType == "" {
			contentType = "application/json"
		}

		resp, err := http.Post(url, contentType, strings.NewReader(data))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("Content-Length: %d\n", len(body))
		fmt.Printf("\nResponse Body:\n%s\n", string(body))
	},
}

var httpPutCmd = &cobra.Command{
	Use:   "http_put <url>",
	Short: "PUT data to HTTP endpoint",
	Long:  `Send PUT request with data to an HTTP endpoint.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		data, _ := cmd.Flags().GetString("data")
		contentType, _ := cmd.Flags().GetString("content-type")

		if data == "" {
			data = "{}"
		}
		if contentType == "" {
			contentType = "application/json"
		}

		req, _ := http.NewRequest("PUT", url, strings.NewReader(data))
		req.Header.Set("Content-Type", contentType)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("\nResponse Body:\n%s\n", string(body))
	},
}

var httpDeleteCmd = &cobra.Command{
	Use:   "http_delete <url>",
	Short: "DELETE request",
	Long:  `Send DELETE request to an HTTP endpoint.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		req, _ := http.NewRequest("DELETE", url, nil)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		if len(body) > 0 {
			fmt.Printf("\nResponse Body:\n%s\n", string(body))
		}
	},
}

var httpPatchCmd = &cobra.Command{
	Use:   "http_patch <url>",
	Short: "PATCH request",
	Long:  `Send PATCH request with data to an HTTP endpoint.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		data, _ := cmd.Flags().GetString("data")
		contentType, _ := cmd.Flags().GetString("content-type")

		if data == "" {
			data = "{}"
		}
		if contentType == "" {
			contentType = "application/json"
		}

		req, _ := http.NewRequest("PATCH", url, strings.NewReader(data))
		req.Header.Set("Content-Type", contentType)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("\nResponse Body:\n%s\n", string(body))
	},
}

var http2TestCmd = &cobra.Command{
	Use:   "http2_test <url>",
	Short: "Test HTTP/2 support",
	Long:  `Check if a server supports HTTP/2 protocol.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			},
		}

		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("Proto:  %s\n", resp.Proto)

		nativeTls := resp.TLS != nil
		fmt.Printf("HTTP/2: %v\n", resp.Proto == "HTTP/2" || nativeTls)
		if resp.TLS != nil {
			fmt.Printf("TLS Version: %d.%d\n", resp.TLS.Version>>8, resp.TLS.Version&0xFF)
			fmt.Printf("Cipher Suite: %s\n", tls.CipherSuiteName(resp.TLS.CipherSuite))
		}
	},
}

var httpHeadersCmd = &cobra.Command{
	Use:   "http_headers <url>",
	Short: "Show HTTP headers",
	Long:  `Display all HTTP response headers from a URL.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		resp, err := http.Head(url)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("URL: %s\n\n", url)
		fmt.Println("=== Response Headers ===")
		for k, v := range resp.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ", "))
		}
	},
}

func init() {
	httpPostCmd.Flags().StringP("data", "d", "", "POST data")
	httpPostCmd.Flags().StringP("content-type", "t", "", "Content-Type header")

	httpPutCmd.Flags().StringP("data", "d", "", "PUT data")
	httpPutCmd.Flags().StringP("content-type", "t", "", "Content-Type header")

	httpPatchCmd.Flags().StringP("data", "d", "", "PATCH data")
	httpPatchCmd.Flags().StringP("content-type", "t", "", "Content-Type header")
}