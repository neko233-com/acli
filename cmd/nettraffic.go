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

var nettrafficCmd = &cobra.Command{
	Use:   "nettraffic",
	Short: "Show network traffic",
	Long:  `Display network interface traffic statistics.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			cmdExec := exec.Command("netstat", "-e")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "linux":
			cmdExec := exec.Command("cat", "/proc/net/dev")
			output, _ := cmdExec.Output()
			printNetDev(string(output))
		case "darwin":
			cmdExec := exec.Command("netstat", "-i")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}

var bandwidthCmd = &cobra.Command{
	Use:   "bandwidth",
	Short: "Measure bandwidth",
	Long:  `Measure network bandwidth by downloading/uploading test data.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetURL, _ := cmd.Flags().GetString("url")
		if targetURL == "" {
			targetURL = "http://speedtest.tele2.net/10MB.zip"
		}

		fmt.Println("Measuring download bandwidth...")
		testBandwidthDownload(targetURL)
	},
}

func init() {
	bandwidthCmd.Flags().String("url", "", "URL to download for bandwidth test")
}

func printNetDev(output string) {
	lines := strings.Split(output, "\n")
	fmt.Println("\n=== Network Interface Traffic ===")
	fmt.Printf("%-10s %15s %15s %15s %15s\n", "Interface", "RX Bytes", "TX Bytes", "RX Packets", "TX Packets")
	fmt.Println(strings.Repeat("-", 70))

	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 10 {
			iface := strings.Trim(fields[0], ":")
			rxBytes := fields[1]
			txBytes := fields[9]
			fmt.Printf("%-10s %15s %15s %15s %15s\n", iface, rxBytes, txBytes, fields[2], fields[10])
		}
	}
}

func testBandwidthDownload(url string) {
	client := &http.Client{Timeout: 60 * time.Second}
	start := time.Now()

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var total int64
	buf := make([]byte, 32*1024)
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
	mbps := (mb * 8) / duration.Seconds()

	fmt.Printf("\nDownloaded: %.2f MB\n", mb)
	fmt.Printf("Duration:   %v\n", duration)
	fmt.Printf("Speed:      %.2f Mbps\n", mbps)
}
