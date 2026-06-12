package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type portEntry struct {
	Protocol string `json:"protocol"`
	Local    string `json:"local"`
	Remote   string `json:"remote,omitempty"`
	State    string `json:"state,omitempty"`
	PID      string `json:"pid,omitempty"`
	Process  string `json:"process,omitempty"`
}

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Show structured network ports",
	Long:  `Show listening/active ports as human-readable table or JSON for agents.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		entries, err := collectPorts()
		if err != nil {
			exitErr(err)
		}
		if jsonOut {
			data, _ := json.MarshalIndent(entries, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("%-6s %-28s %-28s %-14s %-8s %s\n", "PROTO", "LOCAL", "REMOTE", "STATE", "PID", "PROCESS")
		for _, e := range entries {
			fmt.Printf("%-6s %-28s %-28s %-14s %-8s %s\n", e.Protocol, e.Local, e.Remote, e.State, e.PID, e.Process)
		}
	},
}

func init() {
	portsCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
}

func collectPorts() ([]portEntry, error) {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("netstat", "-ano").Output()
		if err != nil {
			return nil, err
		}
		return parseWindowsNetstat(string(out)), nil
	case "linux":
		out, err := exec.Command("ss", "-tunap").Output()
		if err != nil {
			return nil, err
		}
		return parseSSPorts(string(out)), nil
	case "darwin":
		out, err := exec.Command("lsof", "-nP", "-iTCP", "-iUDP").Output()
		if err != nil {
			return nil, err
		}
		return parseLsofPorts(string(out)), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func parseWindowsNetstat(value string) []portEntry {
	entries := []portEntry{}
	for _, line := range strings.Split(value, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		proto := strings.ToLower(fields[0])
		if proto != "tcp" && proto != "udp" {
			continue
		}
		entry := portEntry{Protocol: proto, Local: fields[1]}
		if proto == "tcp" && len(fields) >= 5 {
			entry.Remote = fields[2]
			entry.State = fields[3]
			entry.PID = fields[4]
		} else if proto == "udp" && len(fields) >= 4 {
			entry.Remote = fields[2]
			entry.PID = fields[3]
		}
		entries = append(entries, entry)
	}
	return entries
}

func parseSSPorts(value string) []portEntry {
	entries := []portEntry{}
	for _, line := range strings.Split(value, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || strings.HasPrefix(line, "Netid") {
			continue
		}
		entry := portEntry{Protocol: fields[0], State: fields[1], Local: fields[4]}
		if len(fields) > 5 {
			entry.Remote = fields[5]
		}
		if idx := strings.Index(line, "pid="); idx >= 0 {
			rest := line[idx+4:]
			end := strings.IndexAny(rest, ",)")
			if end > 0 {
				entry.PID = rest[:end]
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func parseLsofPorts(value string) []portEntry {
	entries := []portEntry{}
	for _, line := range strings.Split(value, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 9 || fields[0] == "COMMAND" {
			continue
		}
		entry := portEntry{Process: fields[0], PID: fields[1], Protocol: strings.ToLower(fields[7]), Local: fields[8]}
		if len(fields) > 9 {
			entry.State = strings.Join(fields[9:], " ")
		}
		entries = append(entries, entry)
	}
	return entries
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
