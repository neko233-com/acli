package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List processes",
	Long:  `Display a list of running processes with PID and name.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cmdExec *exec.Cmd
		var err error

		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("tasklist")
		case "linux":
			cmdExec = exec.Command("ps", "-eo", "pid,ppid,%cpu,%mem,comm")
		case "darwin":
			cmdExec = exec.Command("ps", "-eo", "pid,ppid,%cpu,%mem,comm")
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		err = cmdExec.Run()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}
