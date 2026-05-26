package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "Show network connections",
	Long:  `Display active network connections (TCP/UDP).`,
	Run: func(cmd *cobra.Command, args []string) {
		var cmdExec *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("netstat", "-ano")
		case "linux":
			cmdExec = exec.Command("ss", "-tunap")
		case "darwin":
			cmdExec = exec.Command("netstat", "-an")
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		cmdExec.Run()
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect [host]",
	Short: "Test connectivity to a host",
	Long:  `Test TCP connection to a host on common ports.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		ports := []int{80, 443, 22, 3389, 8080}

		fmt.Printf("Testing connectivity to %s...\n\n", host)

		for _, port := range ports {
			addr := fmt.Sprintf("%s:%d", host, port)
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("  Port %d: CLOSED\n", port)
				continue
			}
			conn.Close()
			fmt.Printf("  Port %d: OPEN\n", port)
		}
	},
}