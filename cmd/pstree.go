package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var pstreeCmd = &cobra.Command{
	Use:   "pstree",
	Short: "Show process tree",
	Long:  `Display running processes in a tree view showing parent-child relationships.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cmdExec *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("wmic", "process", "get", "ProcessId,Name,ParentProcessId")
		case "linux":
			cmdExec = exec.Command("ps", "-eo", "pid,ppid,comm")
		case "darwin":
			cmdExec = exec.Command("ps", "-eo", "pid,ppid,comm")
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		cmdExec.Run()
	},
}

var pssearchCmd = &cobra.Command{
	Use:   "pssearch <name>",
	Short: "Search processes by name",
	Long:  `Search for processes matching the given name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.ToLower(args[0])

		var cmdExec *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("tasklist", "/FO", "LIST", "/FI", fmt.Sprintf("IMAGENAME eq %s*", name))
		case "linux":
			cmdExec = exec.Command("ps", "-eo", "pid,ppid,%cpu,%mem,comm", "--no-header")
		case "darwin":
			cmdExec = exec.Command("ps", "-eo", "pid,ppid,%cpu,%mem,comm", "--no-header")
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		output, err := cmdExec.Output()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		lines := strings.Split(string(output), "\n")
		fmt.Println("PID     PPID    CPU%    MEM%    COMMAND")
		for _, line := range lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, name) {
				parts := strings.Fields(line)
				if len(parts) >= 5 {
					fmt.Printf("%-8s%-8s%-8s%-8s%s\n", parts[0], parts[1], parts[2], parts[3], strings.Join(parts[4:], " "))
				} else if len(parts) > 0 {
					fmt.Println(line)
				}
			}
		}
	},
}

var psportsCmd = &cobra.Command{
	Use:   "psports [port]",
	Short: "Show processes using a port",
	Long:  `Display process using specified port (or all processes with network connections if no port specified).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port := args[0]

		switch runtime.GOOS {
		case "windows":
			cmdExec := exec.Command("netstat", "-ano")
			output, err := cmdExec.Output()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			lines := strings.Split(string(output), "\n")
			fmt.Printf("Showing processes for port %s:\n\n", port)
			for _, line := range lines {
				if strings.Contains(line, ":"+port) || strings.Contains(line, "LISTENING") {
					fmt.Println(line)
				}
			}
		case "linux":
			cmdExec := exec.Command("ss", "-tunap", "|", "grep", ":"+port)
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "darwin":
			cmdExec := exec.Command("lsof", "-i", ":"+port)
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}
