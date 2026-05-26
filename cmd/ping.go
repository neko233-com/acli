package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping <host>",
	Short: "Ping a host",
	Long:  `Send ICMP echo requests to test connectivity and measure latency.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		count, _ := cmd.Flags().GetInt("count")
		timeout := time.Duration(5) * time.Second

		fmt.Printf("PING %s (%s)\n\n", host, host)

		var pingCmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			pingCmd = exec.Command("ping", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeout.Milliseconds()), host)
		case "linux":
			pingCmd = exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", "5", host)
		case "darwin":
			pingCmd = exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", "5", host)
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		pingCmd.Stdout = os.Stdout
		pingCmd.Stderr = os.Stderr
		pingCmd.Run()
	},
}

func init() {
	pingCmd.Flags().IntP("count", "c", 4, "Number of packets to send")
}