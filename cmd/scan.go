package cmd

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <host>",
	Short: "Scan common ports on a host",
	Long:  `Scan specified range of common ports to check which are open.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		startPort, _ := cmd.Flags().GetInt("start")
		endPort, _ := cmd.Flags().GetInt("end")
		threads, _ := cmd.Flags().GetInt("threads")

		ports := make([]int, 0)
		for p := startPort; p <= endPort; p++ {
			ports = append(ports, p)
		}

		fmt.Printf("Scanning %s ports %d-%d (%d threads)...\n\n", host, startPort, endPort, threads)

		results := make(chan int, len(ports))
		var wg sync.WaitGroup

		sem := make(chan struct{}, threads)
		for _, port := range ports {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				addr := net.JoinHostPort(host, strconv.Itoa(p))
				conn, err := net.DialTimeout("tcp", addr, time.Second)
				if err == nil {
					conn.Close()
					results <- p
				}
			}(port)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		openPorts := make([]int, 0)
		for port := range results {
			openPorts = append(openPorts, port)
		}

		if len(openPorts) == 0 {
			fmt.Println("No open ports found.")
			return
		}

		fmt.Println("Open ports:")
		serviceName := map[int]string{
			21:   "FTP",
			22:   "SSH",
			23:   "Telnet",
			25:   "SMTP",
			53:   "DNS",
			80:   "HTTP",
			110:  "POP3",
			143:  "IMAP",
			443:  "HTTPS",
			3306: "MySQL",
			3389: "RDP",
			5432: "PostgreSQL",
			6379: "Redis",
			8080: "HTTP-Alt",
			8443: "HTTPS-Alt",
		}

		for _, port := range openPorts {
			name := serviceName[port]
			if name == "" {
				name = "Unknown"
			}
			fmt.Printf("  %d (%s)\n", port, name)
		}
		fmt.Printf("\nFound %d open ports.\n", len(openPorts))
	},
}

func init() {
	scanCmd.Flags().Int("start", 1, "Start port")
	scanCmd.Flags().Int("end", 1024, "End port")
	scanCmd.Flags().Int("threads", 100, "Number of concurrent threads")
}
