package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor [interval]",
	Short: "Process and system monitor (htop-like)",
	Long:  `Real-time process and system resource monitor.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interval := 2
		if len(args) > 0 {
			interval, _ = strconv.Atoi(args[0])
		}

		if runtime.GOOS == "windows" {
			runWindowsMonitor(interval)
		} else {
			runUnixMonitor(interval)
		}
	},
}

func runWindowsMonitor(interval int) {
	for {
		clearScreen()
		printHeader()

		cmdExec := exec.Command("tasklist", "/FO", "LIST", "/NH")
		output, _ := cmdExec.Output()
		lines := strings.Split(string(output), "\n")

		fmt.Println("\n=== Top Processes by Memory ===")
		fmt.Println("PID     MEM%    CPU%   NAME")
		fmt.Println(strings.Repeat("-", 50))

		processes := parseTasklist(lines)
		for i, p := range processes {
			if i >= 15 {
				break
			}
			fmt.Printf("%-8s %-6s %-6s %s\n", p.pid, p.mem, p.cpu, p.name)
		}

		cmdExec = exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/format:list")
		output, _ = cmdExec.Output()
		parseMemInfo(string(output))

		fmt.Printf("\nRefresh in %ds (Ctrl+C to exit)\n", interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func runUnixMonitor(interval int) {
	for {
		clearScreen()
		printHeader()

		cmdExec := exec.Command("ps", "aux", "--sort=-%mem")
		output, _ := cmdExec.Output()
		lines := strings.Split(string(output), "\n")

		fmt.Println("\n=== Top Processes by Memory ===")
		fmt.Println("PID     MEM%    CPU%   COMMAND")
		fmt.Println(strings.Repeat("-", 55))

		for i, line := range lines[1:] {
			parts := strings.Fields(line)
			if len(parts) >= 11 {
				pid := parts[1]
				cpu := parts[2]
				mem := parts[3]
				comm := parts[10]
				fmt.Printf("%-8s %-6s %-6s %s\n", pid, mem, cpu, comm)
				if i >= 14 {
					break
				}
			}
		}

		cmdExec = exec.Command("free", "-b")
		output, _ = cmdExec.Output()
		parseUnixMem(string(output))

		cmdExec = exec.Command("uptime")
		output, _ = cmdExec.Output()
		fmt.Printf("\n%s", strings.TrimSpace(string(output)))
		fmt.Printf("\nRefresh in %ds (Ctrl+C to exit)\n", interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		exec.Command("cmd", "/c", "cls").Run()
	default:
		exec.Command("clear").Run()
	}
}

func printHeader() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    UNICLI MONITOR                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("Platform: %s/%s | Time: %s\n", runtime.GOOS, runtime.GOARCH, time.Now().Format("2006-01-02 15:04:05"))
}

type proc struct {
	pid  string
	mem  string
	cpu  string
	name string
}

func parseTasklist(lines []string) []proc {
	var processes []proc
	for _, line := range lines {
		if strings.Contains(line, "Image Name:") {
			p := proc{}
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				p.name = strings.TrimSpace(parts[1])
			}
			p.name = strings.TrimSuffix(p.name, ".exe")
			processes = append(processes, p)
		}
	}
	return processes
}

func parseMemInfo(output string) {
	fmt.Println("\n=== Memory ===")
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "=") {
			fmt.Println(strings.TrimSpace(line))
		}
	}
}

func parseUnixMem(output string) {
	fmt.Println("\n=== Memory ===")
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			fmt.Println(line)
		}
	}
}