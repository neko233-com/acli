package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var netstatCmd = &cobra.Command{
	Use:   "netstat",
	Short: "Show detailed network statistics",
	Long:  `Display detailed network statistics including TCP, UDP, and ICMP connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cmdExec *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("netstat", "-s")
		case "linux":
			cmdExec = exec.Command("ss", "-s")
		case "darwin":
			cmdExec = exec.Command("netstat", "-s")
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		cmdExec.Run()
	},
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Show listening ports",
	Long:  `Display all ports that are in LISTENING state with process information.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			cmdExec := exec.Command("netstat", "-ano")
			output, err := cmdExec.Output()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			lines := strings.Split(string(output), "\n")
			fmt.Println("Proto  Local Address         PID")
			fmt.Println(strings.Repeat("-", 50))
			for _, line := range lines {
				if strings.Contains(line, "LISTENING") {
					fields := strings.Fields(line)
					if len(fields) >= 5 {
						fmt.Printf("%-8s%-24s%s\n", fields[0], fields[1], fields[len(fields)-1])
					}
				}
			}
		case "linux":
			cmdExec := exec.Command("ss", "-tlnp")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "darwin":
			cmdExec := exec.Command("lsof", "-i", "-P", "-n")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}

var speedtestCmd = &cobra.Command{
	Use:   "speedtest",
	Short: "Network speed test",
	Long:  `Simple network speed test by downloading a test file.`,
	Run: func(cmd *cobra.Command, args []string) {
		testURL := "http://speedtest.tele2.net/1MB.zip"
		fmt.Println("Testing download speed...")

		resp, err := http.Get(testURL)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		buf := make([]byte, 32*1024)
		var total int64
		start := time.Now()

		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				total += int64(n)
			}
			if err != nil {
				break
			}
		}
		duration := time.Since(start)

		mb := float64(total) / (1024 * 1024)
		mbps := mb * 8 / duration.Seconds()

		fmt.Printf("\nDownloaded: %.2f MB\n", mb)
		fmt.Printf("Duration:   %v\n", duration)
		fmt.Printf("Speed:      %.2f Mbps\n", mbps)
	},
}
