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
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid PID")
			os.Exit(1)
		}
		tree, _ := cmd.Flags().GetBool("tree")
		childrenOnly, _ := cmd.Flags().GetBool("children")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		pids := []int{pid}
		if tree || childrenOnly {
			processes, err := collectProcesses()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			descendants := processDescendants(processes, pid)
			pids = []int{}
			for _, child := range descendants {
				pids = append(pids, child.PID)
			}
			if tree {
				pids = append(pids, pid)
			}
		}
		if len(pids) == 0 {
			fmt.Println("No matching processes")
			return
		}
		for i, j := 0, len(pids)-1; i < j; i, j = i+1, j-1 {
			pids[i], pids[j] = pids[j], pids[i]
		}
		if dryRun {
			fmt.Printf("Would kill PIDs: %v\n", pids)
			return
		}

		for _, targetPID := range pids {
			var killExec *exec.Cmd
			switch runtime.GOOS {
			case "windows":
				killExec = exec.Command("taskkill", "/F", "/PID", strconv.Itoa(targetPID))
			case "linux", "darwin":
				killExec = exec.Command("kill", "-9", strconv.Itoa(targetPID))
			default:
				fmt.Println("Unsupported OS")
				os.Exit(1)
			}
			killExec.Stdout = os.Stdout
			killExec.Stderr = os.Stderr
			if err := killExec.Run(); err != nil {
				fmt.Printf("Failed to kill process %d: %v\n", targetPID, err)
			}
		}
	},
}

func init() {
	killCmd.Flags().Bool("tree", false, "Kill process and all descendants")
	killCmd.Flags().Bool("children", false, "Kill only descendants")
	killCmd.Flags().Bool("dry-run", false, "Print target PIDs without killing")
}
