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

var sysinfoCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "Show system information",
	Long:  `Display OS name, version, hostname, uptime, and platform details.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cmdExec *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmdExec = exec.Command("systeminfo")
		case "linux":
			cmdExec = exec.Command("uname", "-a")
		case "darwin":
			cmdExec = exec.Command("sw_vers")
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		cmdExec.Run()
	},
}

var memCmd = &cobra.Command{
	Use:   "mem",
	Short: "Show memory usage",
	Long:  `Display physical and virtual memory usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			cmdExec := exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize,Caption")
			output, err := cmdExec.Output()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 3 {
					total, _ := strconv.ParseUint(fields[1], 10, 64)
					free, _ := strconv.ParseUint(fields[0], 10, 64)
					used := total - free
					fmt.Printf("OS: %s\n", strings.Join(fields[2:], " "))
					fmt.Printf("Total: %s\n", formatBytes(total*1024))
					fmt.Printf("Used:  %s (%.1f%%)\n", formatBytes(used*1024), float64(used)/float64(total)*100)
					fmt.Printf("Free:  %s (%.1f%%)\n", formatBytes(free*1024), float64(free)/float64(total)*100)
				}
			}
		case "linux":
			cmdExec := exec.Command("free", "-b")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "darwin":
			cmdExec := exec.Command("vm_stat")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Show CPU usage",
	Long:  `Display CPU model, core count, and current usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			cmdExec := exec.Command("wmic", "cpu", "get", "Name,NumberOfCores,NumberOfLogicalProcessors,LoadPercentage")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "linux":
			cmdExec := exec.Command("top", "-bn1", "|", "grep", "Cpu(s)")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "darwin":
			cmdExec := exec.Command("sysctl", "-n", "hw.model", "hw.ncpu")
			output, err := cmdExec.Output()
			if err == nil {
				fields := strings.Split(strings.TrimSpace(string(output)), "\n")
				if len(fields) >= 2 {
					fmt.Printf("Model:  %s\n", fields[0])
					fmt.Printf("Cores:  %s\n", fields[1])
				}
			}
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Show disk usage",
	Long:  `Display disk space usage for all mounted partitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			cmdExec := exec.Command("wmic", "logicaldisk", "get", "DeviceID,Size,FreeSpace,Caption")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "linux":
			cmdExec := exec.Command("df", "-h")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		case "darwin":
			cmdExec := exec.Command("df", "-h")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
