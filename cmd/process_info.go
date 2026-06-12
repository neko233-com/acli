package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type procInfo struct {
	PID      int        `json:"pid"`
	PPID     int        `json:"ppid"`
	Name     string     `json:"name"`
	Command  string     `json:"command,omitempty"`
	CPU      string     `json:"cpu,omitempty"`
	Memory   string     `json:"memory,omitempty"`
	Children []procInfo `json:"children,omitempty"`
}

func collectProcesses() ([]procInfo, error) {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_Process | ForEach-Object { \"$($_.ProcessId)|$($_.ParentProcessId)|$($_.Name)|$($_.WorkingSetSize)|$($_.CommandLine)\" }").Output()
		if err != nil {
			return nil, err
		}
		return parsePipeProcesses(string(out)), nil
	default:
		out, err := exec.Command("ps", "-eo", "pid=,ppid=,pcpu=,pmem=,comm=,args=").Output()
		if err != nil {
			return nil, err
		}
		return parsePSProcesses(string(out)), nil
	}
}

func parsePipeProcesses(value string) []procInfo {
	result := []procInfo{}
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 5)
		if len(parts) < 4 {
			continue
		}
		pid, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		ppid, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		command := ""
		if len(parts) > 4 {
			command = strings.TrimSpace(parts[4])
		}
		result = append(result, procInfo{PID: pid, PPID: ppid, Name: strings.TrimSpace(parts[2]), Memory: strings.TrimSpace(parts[3]), Command: command})
	}
	return result
}

func parsePSProcesses(value string) []procInfo {
	result := []procInfo{}
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		ppid, _ := strconv.Atoi(fields[1])
		command := ""
		if len(fields) > 5 {
			command = strings.Join(fields[5:], " ")
		}
		result = append(result, procInfo{PID: pid, PPID: ppid, CPU: fields[2], Memory: fields[3], Name: fields[4], Command: command})
	}
	return result
}

func findProcessTree(processes []procInfo, rootPID int) (procInfo, bool) {
	byParent := map[int][]procInfo{}
	byPID := map[int]procInfo{}
	for _, p := range processes {
		byPID[p.PID] = p
		byParent[p.PPID] = append(byParent[p.PPID], p)
	}
	root, ok := byPID[rootPID]
	if !ok {
		return procInfo{}, false
	}
	return attachChildren(root, byParent), true
}

func attachChildren(root procInfo, byParent map[int][]procInfo) procInfo {
	children := byParent[root.PID]
	sort.Slice(children, func(i, j int) bool { return children[i].PID < children[j].PID })
	for _, child := range children {
		root.Children = append(root.Children, attachChildren(child, byParent))
	}
	return root
}

func processDescendants(processes []procInfo, pid int) []procInfo {
	tree, ok := findProcessTree(processes, pid)
	if !ok {
		return nil
	}
	result := []procInfo{}
	var walk func(procInfo)
	walk = func(p procInfo) {
		for _, child := range p.Children {
			result = append(result, child)
			walk(child)
		}
	}
	walk(tree)
	return result
}

func printProcessTable(processes []procInfo) {
	fmt.Printf("%-8s %-8s %-8s %-8s %s\n", "PID", "PPID", "CPU", "MEM", "NAME")
	for _, p := range processes {
		fmt.Printf("%-8d %-8d %-8s %-8s %s\n", p.PID, p.PPID, p.CPU, p.Memory, p.Name)
	}
}

func printProcessJSON(value any) {
	data, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(data))
}
