package cmd

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var base64EncodeCmd = &cobra.Command{
	Use:   "base64_encode <string>",
	Short: "Encode string to Base64",
	Long:  `Encode a string to Base64 format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data := []byte(args[0])
		encoded := base64.StdEncoding.EncodeToString(data)
		fmt.Printf("Input:  %s\n", args[0])
		fmt.Printf("Output: %s\n", encoded)
	},
}

var base64DecodeCmd = &cobra.Command{
	Use:   "base64_decode <string>",
	Short: "Decode Base64 string",
	Long:  `Decode a Base64 string to plain text.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		decoded, err := base64.StdEncoding.DecodeString(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Input:  %s\n", args[0])
		fmt.Printf("Output: %s\n", string(decoded))
	},
}

var urlEncodeCmd = &cobra.Command{
	Use:   "url_encode <string>",
	Short: "URL encode string",
	Long:  `Encode a string for use in URLs.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		encoded := url.QueryEscape(args[0])
		fmt.Printf("Input:  %s\n", args[0])
		fmt.Printf("Output: %s\n", encoded)
	},
}

var urlDecodeCmd = &cobra.Command{
	Use:   "url_decode <string>",
	Short: "URL decode string",
	Long:  `Decode a URL-encoded string.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		decoded, err := url.QueryUnescape(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Input:  %s\n", args[0])
		fmt.Printf("Output: %s\n", decoded)
	},
}

var hashMd5Cmd = &cobra.Command{
	Use:   "hash_md5 <string>",
	Short: "Calculate MD5 hash",
	Long:  `Calculate MD5 hash of a string.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data := []byte(args[0])
		hash := md5.Sum(data)
		fmt.Printf("Input: %s\n", args[0])
		fmt.Printf("MD5:   %x\n", hash)
	},
}

var hashSha256Cmd = &cobra.Command{
	Use:   "hash_sha256 <string>",
	Short: "Calculate SHA256 hash",
	Long:  `Calculate SHA256 hash of a string.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data := []byte(args[0])
		hash := sha256.Sum256(data)
		fmt.Printf("Input:  %s\n", args[0])
		fmt.Printf("SHA256: %x\n", hash)
	},
}

var uuidGenCmd = &cobra.Command{
	Use:   "uuid_generate",
	Short: "Generate UUID",
	Long:  `Generate a random UUID v4.`,
	Run: func(cmd *cobra.Command, args []string) {
		uuid := uuid.New().String()
		fmt.Printf("UUID v4: %s\n", uuid)
	},
}

var jsonPrettyCmd = &cobra.Command{
	Use:   "json_pretty <file>",
	Short: "Pretty print JSON file",
	Long:  `Format and display JSON file with syntax highlighting.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			fmt.Printf("Invalid JSON: %v\n", err)
			os.Exit(1)
		}

		prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
		fmt.Printf("%s\n", string(prettyJSON))
	},
}

var jsonValidateCmd = &cobra.Command{
	Use:   "json_validate <file>",
	Short: "Validate JSON file",
	Long:  `Check if a file contains valid JSON.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		if json.Valid(data) {
			fmt.Println("Valid JSON")
		} else {
			fmt.Println("Invalid JSON")
			var jsonData interface{}
			if err := json.Unmarshal(data, &jsonData); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			os.Exit(1)
		}
	},
}

var randomStringCmd = &cobra.Command{
	Use:   "random_string [length]",
	Short: "Generate random string",
	Long:  `Generate a random string of specified length.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length := 32
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &length)
		}

		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		rand.Seed(time.Now().UnixNano())

		result := make([]byte, length)
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}

		fmt.Printf("Random string (%d chars): %s\n", length, string(result))
	},
}

var crtLookupCmd = &cobra.Command{
	Use:   "cert_lookup <host>",
	Short: "Certificate lookup",
	Long:  `Lookup SSL certificate info for a host.`,
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

		state := conn.ConnectionState()
		cert := state.PeerCertificates[0]

		fmt.Printf("Certificate for %s:%d\n\n", host, port)
		fmt.Printf("Subject:   %s\n", cert.Subject.CommonName)
		fmt.Printf("Issuer:    %s\n", cert.Issuer.CommonName)
		fmt.Printf("Valid:     %s to %s\n", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
		fmt.Printf("Algorithm: %s\n", cert.SignatureAlgorithm.String())

		if len(cert.DNSNames) > 0 {
			fmt.Printf("SANs:      %s\n", strings.Join(cert.DNSNames, ", "))
		}
	},
}

var hexDumpCmd = &cobra.Command{
	Use:   "hex_dump <file>",
	Short: "Display hex dump",
	Long:  `Display hexadecimal dump of a file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		for i := 0; i < len(data); i += 16 {
			fmt.Printf("%08x  ", i)

			for j := 0; j < 16; j++ {
				if i+j < len(data) {
					fmt.Printf("%02x ", data[i+j])
				} else {
					fmt.Print("   ")
				}
				if j == 7 {
					fmt.Print(" ")
				}
			}

			fmt.Print(" |")
			for j := 0; j < 16 && i+j < len(data); j++ {
				c := data[i+j]
				if c >= 32 && c <= 126 {
					fmt.Printf("%c", c)
				} else {
					fmt.Print(".")
				}
			}
			fmt.Println("|")
		}
	},
}