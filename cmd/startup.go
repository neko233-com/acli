package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type startupEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
	Command string `json:"command,omitempty"`
}

var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "Manage user startup entries",
	Long:  `List, add, remove, enable, and disable per-user startup entries across Windows, Linux, and macOS.`,
}

var startupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List startup entries",
	Run: func(cmd *cobra.Command, args []string) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		entries, err := listStartupEntries()
		if err != nil {
			exitErr(err)
		}
		if jsonOut {
			data, _ := json.MarshalIndent(entries, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("%-24s %-8s %s\n", "NAME", "ENABLED", "PATH")
		for _, e := range entries {
			fmt.Printf("%-24s %-8v %s\n", e.Name, e.Enabled, e.Path)
		}
	},
}

var startupAddCmd = &cobra.Command{
	Use:   "add <name> -- <command>",
	Short: "Add startup entry",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if len(args) > 1 && args[1] == "--" {
			args = append(args[:1], args[2:]...)
		}
		if len(args) < 2 {
			exitErr(fmt.Errorf("missing command"))
		}
		command := strings.Join(args[1:], " ")
		path, err := writeStartupEntry(name, command, true)
		if err != nil {
			exitErr(err)
		}
		fmt.Printf("Added startup entry: %s\n", path)
	},
}

var startupRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove startup entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := removeStartupEntry(args[0]); err != nil {
			exitErr(err)
		}
		fmt.Printf("Removed startup entry: %s\n", args[0])
	},
}

var startupEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable startup entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := setStartupEnabled(args[0], true); err != nil {
			exitErr(err)
		}
		fmt.Printf("Enabled startup entry: %s\n", args[0])
	},
}

var startupDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable startup entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := setStartupEnabled(args[0], false); err != nil {
			exitErr(err)
		}
		fmt.Printf("Disabled startup entry: %s\n", args[0])
	},
}

func init() {
	startupListCmd.Flags().Bool("json", false, "Output JSON for agents/scripts")
	startupCmd.AddCommand(startupListCmd, startupAddCmd, startupRemoveCmd, startupEnableCmd, startupDisableCmd)
}

func startupDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), nil
	case "darwin":
		return filepath.Join(home, "Library", "LaunchAgents"), nil
	default:
		return filepath.Join(home, ".config", "autostart"), nil
	}
}

func startupPath(name string, enabled bool) (string, error) {
	dir, err := startupDir()
	if err != nil {
		return "", err
	}
	ext := ".desktop"
	if runtime.GOOS == "windows" {
		ext = ".cmd"
	} else if runtime.GOOS == "darwin" {
		ext = ".plist"
	}
	if !enabled {
		ext += ".disabled"
	}
	return filepath.Join(dir, safeStartupName(name)+ext), nil
}

func safeStartupName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return replacer.Replace(name)
}

func listStartupEntries() ([]startupEntry, error) {
	dir, err := startupDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := []startupEntry{}
	for _, item := range items {
		name := item.Name()
		if !strings.HasSuffix(name, ".cmd") && !strings.HasSuffix(name, ".desktop") && !strings.HasSuffix(name, ".plist") && !strings.HasSuffix(name, ".disabled") {
			continue
		}
		path := filepath.Join(dir, name)
		enabled := !strings.HasSuffix(name, ".disabled")
		command := readStartupCommand(path)
		entries = append(entries, startupEntry{Name: strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".disabled"), ".desktop"), ".plist"), ".cmd"), Path: path, Enabled: enabled, Command: command})
	}
	return entries, nil
}

func writeStartupEntry(name, command string, enabled bool) (string, error) {
	path, err := startupPath(name, enabled)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	var content string
	switch runtime.GOOS {
	case "windows":
		content = "@echo off\r\n" + command + "\r\n"
	case "darwin":
		label := "com.unicli." + safeStartupName(name)
		content = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>%s</string>
<key>ProgramArguments</key><array><string>/bin/sh</string><string>-lc</string><string>%s</string></array>
<key>RunAtLoad</key><true/>
</dict></plist>
`, label, command)
	default:
		content = fmt.Sprintf("[Desktop Entry]\nType=Application\nName=%s\nExec=sh -lc '%s'\nX-GNOME-Autostart-enabled=true\n", name, strings.ReplaceAll(command, "'", "'\\''"))
	}
	return path, os.WriteFile(path, []byte(content), 0755)
}

func removeStartupEntry(name string) error {
	for _, enabled := range []bool{true, false} {
		path, err := startupPath(name, enabled)
		if err == nil {
			_ = os.Remove(path)
		}
	}
	return nil
}

func setStartupEnabled(name string, enabled bool) error {
	oldPath, err := startupPath(name, !enabled)
	if err != nil {
		return err
	}
	newPath, err := startupPath(name, enabled)
	if err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func readStartupCommand(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
