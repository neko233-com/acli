package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Show firewall rules",
	Long:  `Display current firewall rules (requires admin/root).`,
	Run: func(cmd *cobra.Command, args []string) {
		showAll, _ := cmd.Flags().GetBool("all")

		switch runtime.GOOS {
		case "windows":
			showWindowsFirewall(showAll)
		case "linux":
			showLinuxFirewall(showAll)
		case "darwin":
			showMacFirewall(showAll)
		default:
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}
	},
}

func showWindowsFirewall(showAll bool) {
	fmt.Println("=== Windows Firewall Rules ===\n")

	_ = showAll // used for future filtering

	cmdExec := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all")
	output, err := cmdExec.Output()
	if err != nil {
		fmt.Printf("Error (try running as Administrator): %v\n", err)
		return
	}

	lines := strings.Split(string(output), "\n")
	enabled := 0
	disabled := 0

	for _, line := range lines {
		if strings.Contains(line, "Rule Name:") {
			name := strings.TrimSpace(strings.Split(line, ":")[1])
			fmt.Printf("\n[%s]\n", name)
		}
		if strings.Contains(line, "Enabled:") {
			enabledStr := strings.TrimSpace(strings.Split(line, ":")[1])
			fmt.Printf("  Status: %s\n", enabledStr)
			if enabledStr == "Yes" {
				enabled++
			} else {
				disabled++
			}
		}
		if strings.Contains(line, "Direction:") {
			fmt.Printf("  %s\n", strings.TrimSpace(line))
		}
		if strings.Contains(line, "Action:") {
			fmt.Printf("  %s\n", strings.TrimSpace(line))
		}
		if strings.Contains(line, "Protocol:") {
			fmt.Printf("  %s\n", strings.TrimSpace(line))
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Enabled:  %d\n", enabled)
	fmt.Printf("Disabled: %d\n", disabled)
}

func showLinuxFirewall(showAll bool) {
	fmt.Println("=== Linux Firewall Rules ===\n")

	iptablesExists := isCommandAvailable("iptables")

	if iptablesExists {
		cmdExec := exec.Command("iptables", "-L", "-n", "-v")
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		cmdExec.Run()
	} else {
		fmt.Println("iptables not found. Checking ufw...")
		if isCommandAvailable("ufw") {
			cmdExec := exec.Command("ufw", "status", "verbose")
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			cmdExec.Run()
		} else {
			fmt.Println("No firewall tool found (tried: iptables, ufw)")
		}
	}
}

func showMacFirewall(showAll bool) {
	fmt.Println("=== macOS Firewall Rules ===\n")

	if isCommandAvailable("pfctl") {
		cmdExec := exec.Command("pfctl", "-s", "rules")
		output, _ := cmdExec.Output()
		fmt.Println(string(output))
	} else {
		cmdExec := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--listall")
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		cmdExec.Run()
	}

	fmt.Println("\nNote: Some rules require admin privileges to view.")
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func init() {
	firewallCmd.Flags().BoolP("all", "a", false, "Show all rules (including disabled)")
}