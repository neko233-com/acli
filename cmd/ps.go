package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List processes",
	Long:  `Display a list of running processes with PID and name.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		filter, _ := cmd.Flags().GetString("filter")
		if jsonOut || filter != "" {
			processes, err := collectProcesses()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if filter != "" {
				filter = strings.ToLower(filter)
				kept := []procInfo{}
				for _, p := range processes {
					if strings.Contains(strings.ToLower(p.Name), filter) || strings.Contains(strings.ToLower(p.Command), filter) {
						kept = append(kept, p)
					}
				}
				processes = kept
			}
			if jsonOut {
				printProcessJSON(processes)
				return
			}
			printProcessTable(processes)
			return
		}
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

var procCmd = &cobra.Command{
	Use:   "proc <pid>",
	Short: "Show process detail and children",
	Long:  `Show one process with parent/children dependency details.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid PID")
			os.Exit(1)
		}
		jsonOut, _ := cmd.Flags().GetBool("json")
		processes, err := collectProcesses()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		tree, ok := findProcessTree(processes, pid)
		if !ok {
			fmt.Printf("Process not found: %d\n", pid)
			os.Exit(1)
		}
		if jsonOut {
			printProcessJSON(tree)
			return
		}
		printProcessTable([]procInfo{tree})
		if len(tree.Children) > 0 {
			fmt.Println("Children:")
			printProcessTree(tree.Children, "")
		}
	},
}

func init() {
	psCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
	psCmd.Flags().String("filter", "", "Filter by process name or command")
	procCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
}
