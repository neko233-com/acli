package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill [pid]",
	Short: "Kill a process by PID",
	Long:  `Terminate a process by its process ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid PID")
			os.Exit(1)
		}

		var killExec *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			killExec = exec.Command("taskkill", "/F", "/PID", args[0])
		case "linux", "darwin":
			killExec = exec.Command("kill", "-9", args[0])
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		killExec.Stdout = os.Stdout
		killExec.Stderr = os.Stderr
		if err := killExec.Run(); err != nil {
			fmt.Printf("Failed to kill process: %v\n", err)
		}
	},
}