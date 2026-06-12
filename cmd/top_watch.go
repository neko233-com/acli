package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type topFrame struct {
	Time      string     `json:"time"`
	Hostname  string     `json:"hostname"`
	Processes []procInfo `json:"processes"`
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Agent-friendly process monitor",
	Long:  `Show top processes as a table, JSON frame, or NDJSON stream.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		stream, _ := cmd.Flags().GetBool("stream")
		interval, _ := cmd.Flags().GetDuration("interval")
		limit, _ := cmd.Flags().GetInt("limit")
		sortBy, _ := cmd.Flags().GetString("sort")
		filter, _ := cmd.Flags().GetString("filter")
		if interval <= 0 {
			interval = 2 * time.Second
		}
		for {
			frame, err := collectTopFrame(limit, sortBy, filter)
			if err != nil {
				exitErr(err)
			}
			if jsonOut || stream {
				data, _ := json.Marshal(frame)
				fmt.Println(string(data))
			} else {
				fmt.Printf("Time: %s Host: %s\n", frame.Time, frame.Hostname)
				printProcessTable(frame.Processes)
			}
			if !stream {
				return
			}
			time.Sleep(interval)
		}
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch -- <command>",
	Short: "Run a command repeatedly",
	Long:  `Run a command repeatedly. With --stream, each result is emitted as one JSON frame for agents.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetDuration("interval")
		stream, _ := cmd.Flags().GetBool("stream")
		if interval <= 0 {
			interval = 2 * time.Second
		}
		if len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		if len(args) == 0 {
			exitErr(fmt.Errorf("missing command"))
		}
		for {
			start := time.Now()
			out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
			if stream {
				frame := map[string]any{
					"time":       start.Format(time.RFC3339),
					"command":    args,
					"ok":         err == nil,
					"output":     string(out),
					"elapsed_ms": time.Since(start).Milliseconds(),
				}
				if err != nil {
					frame["error"] = err.Error()
				}
				data, _ := json.Marshal(frame)
				fmt.Println(string(data))
			} else {
				fmt.Printf("=== %s ===\n", start.Format(time.RFC3339))
				fmt.Print(string(out))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			}
			time.Sleep(interval)
		}
	},
}

func init() {
	topCmd.Flags().Bool("json", false, "Output one JSON frame")
	topCmd.Flags().Bool("stream", false, "Output NDJSON frames continuously")
	topCmd.Flags().Duration("interval", 2*time.Second, "Stream interval")
	topCmd.Flags().Int("limit", 20, "Max processes")
	topCmd.Flags().String("sort", "memory", "Sort by memory, cpu, pid, or name")
	topCmd.Flags().String("filter", "", "Filter process name or command")

	watchCmd.Flags().Duration("interval", 2*time.Second, "Run interval")
	watchCmd.Flags().Bool("stream", false, "Output one JSON frame per run")
}

func collectTopFrame(limit int, sortBy, filter string) (topFrame, error) {
	processes, err := collectProcesses()
	if err != nil {
		return topFrame{}, err
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
	sort.Slice(processes, func(i, j int) bool {
		switch sortBy {
		case "cpu":
			return parseFloat(processes[i].CPU) > parseFloat(processes[j].CPU)
		case "pid":
			return processes[i].PID < processes[j].PID
		case "name":
			return processes[i].Name < processes[j].Name
		default:
			return parseFloat(processes[i].Memory) > parseFloat(processes[j].Memory)
		}
	})
	if limit > 0 && len(processes) > limit {
		processes = processes[:limit]
	}
	hostname, _ := os.Hostname()
	return topFrame{Time: time.Now().Format(time.RFC3339), Hostname: hostname, Processes: processes}, nil
}

func parseFloat(value string) float64 {
	var result float64
	fmt.Sscanf(value, "%f", &result)
	return result
}
