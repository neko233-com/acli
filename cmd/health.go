package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type healthResult struct {
	Hostname     string         `json:"hostname"`
	Platform     string         `json:"platform"`
	CPU          map[string]any `json:"cpu,omitempty"`
	Memory       map[string]any `json:"memory,omitempty"`
	Disk         []diskHealth   `json:"disk,omitempty"`
	Load         string         `json:"load,omitempty"`
	TopProcesses []processInfo  `json:"top_processes,omitempty"`
	Ports        []portHealth   `json:"ports,omitempty"`
}

type diskHealth struct {
	Name        string `json:"name"`
	TotalBytes  uint64 `json:"total_bytes,omitempty"`
	FreeBytes   uint64 `json:"free_bytes,omitempty"`
	UsedPercent string `json:"used_percent,omitempty"`
}

type processInfo struct {
	PID  string `json:"pid"`
	CPU  string `json:"cpu,omitempty"`
	Mem  string `json:"mem,omitempty"`
	Name string `json:"name"`
}

type portHealth struct {
	Port  int    `json:"port"`
	Open  bool   `json:"open"`
	Error string `json:"error,omitempty"`
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "One-shot system health summary",
	Long:  `Show CPU, memory, disk, load, top process, and port check summary. Use --json for agent-readable output.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		portValues, _ := cmd.Flags().GetIntSlice("port")
		result := collectHealth(portValues)
		if jsonOut {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Hostname: %s\nPlatform: %s\n", result.Hostname, result.Platform)
		if result.Load != "" {
			fmt.Printf("Load: %s\n", result.Load)
		}
		if len(result.Disk) > 0 {
			fmt.Println("Disk:")
			for _, disk := range result.Disk {
				fmt.Printf("  %s total=%s free=%s used=%s\n", disk.Name, formatBytes(disk.TotalBytes), formatBytes(disk.FreeBytes), disk.UsedPercent)
			}
		}
		if len(result.TopProcesses) > 0 {
			fmt.Println("Top Processes:")
			for _, p := range result.TopProcesses {
				fmt.Printf("  pid=%s cpu=%s mem=%s %s\n", p.PID, p.CPU, p.Mem, p.Name)
			}
		}
		if len(result.Ports) > 0 {
			fmt.Println("Ports:")
			for _, p := range result.Ports {
				fmt.Printf("  %d open=%v %s\n", p.Port, p.Open, p.Error)
			}
		}
	},
}

func init() {
	healthCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
	healthCmd.Flags().IntSlice("port", nil, "TCP port to check on localhost (repeatable)")
}

func collectHealth(ports []int) healthResult {
	hostname, _ := os.Hostname()
	result := healthResult{
		Hostname:     hostname,
		Platform:     runtime.GOOS + "/" + runtime.GOARCH,
		CPU:          collectCPUHealth(),
		Memory:       collectMemoryHealth(),
		Disk:         collectDiskHealth(),
		Load:         collectLoad(),
		TopProcesses: collectTopProcesses(),
	}
	for _, port := range ports {
		result.Ports = append(result.Ports, checkLocalPort(port))
	}
	return result
}

func collectCPUHealth() map[string]any {
	result := map[string]any{"logical_cpus": runtime.NumCPU()}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_Processor | Select-Object -First 1 Name,LoadPercentage | ConvertTo-Json -Compress").Output()
		if err == nil {
			result["raw"] = strings.TrimSpace(string(out))
		}
	}
	return result
}

func collectMemoryHealth() map[string]any {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "$os=Get-CimInstance Win32_OperatingSystem; [pscustomobject]@{total_bytes=$os.TotalVisibleMemorySize*1KB; free_bytes=$os.FreePhysicalMemory*1KB} | ConvertTo-Json -Compress").Output()
		if err == nil {
			return map[string]any{"raw": strings.TrimSpace(string(out))}
		}
	case "linux":
		out, err := exec.Command("sh", "-c", "free -b | awk '/Mem:/ {printf \"total_bytes=%s free_bytes=%s used_bytes=%s\", $2, $4, $3}'").Output()
		if err == nil {
			return map[string]any{"raw": strings.TrimSpace(string(out))}
		}
	}
	return nil
}

func collectDiskHealth() []diskHealth {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_LogicalDisk -Filter \"DriveType=3\" | ForEach-Object { \"$($_.DeviceID)|$($_.Size)|$($_.FreeSpace)\" }").Output()
		if err == nil {
			disks := []diskHealth{}
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				parts := strings.Split(strings.TrimSpace(line), "|")
				if len(parts) != 3 {
					continue
				}
				total, _ := strconv.ParseUint(parts[1], 10, 64)
				free, _ := strconv.ParseUint(parts[2], 10, 64)
				disks = append(disks, diskHealth{Name: parts[0], TotalBytes: total, FreeBytes: free, UsedPercent: percentUsed(total, free)})
			}
			return disks
		}
	default:
		out, err := exec.Command("sh", "-c", "df -kP | tail -n +2").Output()
		if err == nil {
			disks := []diskHealth{}
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				fields := strings.Fields(line)
				if len(fields) < 6 {
					continue
				}
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				free, _ := strconv.ParseUint(fields[3], 10, 64)
				disks = append(disks, diskHealth{Name: fields[5], TotalBytes: total * 1024, FreeBytes: free * 1024, UsedPercent: fields[4]})
			}
			return disks
		}
	}
	return nil
}

func collectLoad() string {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	default:
		out, err := exec.Command("sh", "-c", "uptime").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

func collectTopProcesses() []processInfo {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-Process | Sort-Object WS -Descending | Select-Object -First 5 Id,ProcessName,CPU,WS | ForEach-Object { \"$($_.Id)|$($_.CPU)|$($_.WS)|$($_.ProcessName)\" }").Output()
		if err == nil {
			return parseProcessLines(string(out))
		}
	default:
		out, err := exec.Command("sh", "-c", "ps -eo pid,pcpu,pmem,comm --sort=-%mem | head -n 6 | tail -n +2").Output()
		if err == nil {
			processes := []processInfo{}
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 4 {
					processes = append(processes, processInfo{PID: fields[0], CPU: fields[1], Mem: fields[2], Name: fields[3]})
				}
			}
			return processes
		}
	}
	return nil
}

func parseProcessLines(value string) []processInfo {
	processes := []processInfo{}
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		parts := strings.Split(strings.TrimSpace(line), "|")
		if len(parts) >= 4 {
			processes = append(processes, processInfo{PID: parts[0], CPU: parts[1], Mem: parts[2], Name: parts[3]})
		}
	}
	return processes
}

func checkLocalPort(port int) portHealth {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 2_000_000_000)
	if err != nil {
		return portHealth{Port: port, Open: false, Error: err.Error()}
	}
	conn.Close()
	return portHealth{Port: port, Open: true}
}

func percentUsed(total, free uint64) string {
	if total == 0 {
		return ""
	}
	used := float64(total-free) / float64(total) * 100
	return fmt.Sprintf("%.1f%%", used)
}
