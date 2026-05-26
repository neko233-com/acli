package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var tracerouteCmd = &cobra.Command{
	Use:   "traceroute <host>",
	Short: "Trace route to host",
	Long:  `Show the route packets take to reach a network host.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]

		var traceExec *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			traceExec = exec.Command("tracert", host)
		case "linux":
			traceExec = exec.Command("traceroute", host)
		case "darwin":
			traceExec = exec.Command("traceroute", host)
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		traceExec.Stdout = os.Stdout
		traceExec.Stderr = os.Stderr
		traceExec.Run()
	},
}
