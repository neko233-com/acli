package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [file|profile:/path|service]",
	Short: "Read local, remote, or service logs",
	Long:  `Unified logs command. Reads plain files, remote files over SSH, Linux systemd journal, macOS log stream, or Windows Event Log.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")
		service, _ := cmd.Flags().GetString("service")
		eventLog, _ := cmd.Flags().GetString("eventlog")
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		if target != "" {
			if err := runTailTarget(cmd, target, lines, follow); err != nil {
				exitErr(err)
			}
			return
		}
		if service != "" {
			runServiceLogs(service, lines, follow)
			return
		}
		if runtime.GOOS == "windows" {
			runWindowsEventLogs(eventLog, lines)
			return
		}
		exitErr(fmt.Errorf("pass a file/remote path or --service"))
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	logsCmd.Flags().IntP("lines", "n", 100, "Number of lines/events")
	logsCmd.Flags().String("service", "", "Service name for system logs")
	logsCmd.Flags().String("eventlog", "System", "Windows event log name")
	for _, c := range []*cobra.Command{logsCmd} {
		c.Flags().StringP("user", "u", "", "SSH username for remote path")
		c.Flags().IntP("port", "p", defaultSSHPort, "SSH port for remote path")
		c.Flags().StringP("key", "i", "", "Private key path for remote path")
		c.Flags().String("passphrase", "", "Private key passphrase for remote path")
		c.Flags().String("password", "", "SSH password for remote path")
		c.Flags().Bool("no-bootstrap", false, "Do not install default public key after password login")
		c.Flags().Duration("timeout", 15*time.Second, "SSH connection timeout")
	}
}

func runServiceLogs(service string, lines int, follow bool) {
	switch runtime.GOOS {
	case "linux":
		args := []string{"-u", service, "-n", strconv.Itoa(lines), "--no-pager"}
		if follow {
			args = append(args, "-f")
		}
		runStreamCommand("journalctl", args...)
	case "darwin":
		predicate := fmt.Sprintf("process == %q", service)
		args := []string{"show", "--last", "1h", "--predicate", predicate}
		if follow {
			args = []string{"stream", "--predicate", predicate}
		}
		runStreamCommand("log", args...)
	case "windows":
		runWindowsEventLogs(service, lines)
	default:
		exitErr(fmt.Errorf("unsupported OS: %s", runtime.GOOS))
	}
}

func runWindowsEventLogs(name string, lines int) {
	if strings.TrimSpace(name) == "" {
		name = "System"
	}
	command := fmt.Sprintf("Get-WinEvent -LogName %q -MaxEvents %d | Select-Object TimeCreated,ProviderName,Id,LevelDisplayName,Message | Format-List", name, lines)
	runStreamCommand("powershell", "-NoProfile", "-Command", command)
}

func runStreamCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		exitErr(err)
	}
}
