package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var serviceListCmd = &cobra.Command{
	Use:   "service_list",
	Short: "List services",
	Long:  `List all system services (Windows services or systemd units).`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			listWindowsServices()
		case "linux":
			listSystemdServices()
		case "darwin":
			listLaunchdServices()
		}
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "service_status <name>",
	Short: "Check service status",
	Long:  `Check the status of a specific service.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		switch runtime.GOOS {
		case "windows":
			checkWindowsService(name)
		case "linux":
			checkSystemdService(name)
		case "darwin":
			checkLaunchdService(name)
		}
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "service_start <name>",
	Short: "Start a service",
	Long:  `Start a system service.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		switch runtime.GOOS {
		case "windows":
			runCommand("net", "start", name)
		case "linux":
			runCommand("sudo", "systemctl", "start", name)
		case "darwin":
			runCommand("sudo", "launchctl", "load", "/Library/LaunchDaemons/"+name+".plist")
		}
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "service_stop <name>",
	Short: "Stop a service",
	Long:  `Stop a system service.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		switch runtime.GOOS {
		case "windows":
			runCommand("net", "stop", name)
		case "linux":
			runCommand("sudo", "systemctl", "stop", name)
		case "darwin":
			runCommand("sudo", "launchctl", "unload", "/Library/LaunchDaemons/"+name+".plist")
		}
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "service_restart <name>",
	Short: "Restart a service",
	Long:  `Restart a system service.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		switch runtime.GOOS {
		case "windows":
			runCommand("net", "stop", name)
			runCommand("net", "start", name)
		case "linux":
			runCommand("sudo", "systemctl", "restart", name)
		case "darwin":
			runCommand("sudo", "launchctl", "unload", "/Library/LaunchDaemons/"+name+".plist")
			runCommand("sudo", "launchctl", "load", "/Library/LaunchDaemons/"+name+".plist")
		}
	},
}

func listWindowsServices() {
	cmd := exec.Command("sc", "query", "state=", "all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func listSystemdServices() {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func listLaunchdServices() {
	cmd := exec.Command("launchctl", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func checkWindowsService(name string) {
	cmd := exec.Command("sc", "query", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func checkSystemdService(name string) {
	cmd := exec.Command("systemctl", "status", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func checkLaunchdService(name string) {
	cmd := exec.Command("launchctl", "list", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

var envListCmd = &cobra.Command{
	Use:   "env_list",
	Short: "List environment variables",
	Long:  `Display all environment variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				fmt.Printf("%s=%s\n", parts[0], parts[1])
			}
		}
	},
}

var envGetCmd = &cobra.Command{
	Use:   "env_get <name>",
	Short: "Get environment variable",
	Long:  `Get the value of a specific environment variable.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value := os.Getenv(args[0])
		if value != "" {
			fmt.Printf("%s=%s\n", args[0], value)
		} else {
			fmt.Printf("Environment variable '%s' not found\n", args[0])
		}
	},
}

var uptimeCmd = &cobra.Command{
	Use:   "uptime",
	Short: "Show system uptime",
	Long:  `Display system uptime and load average.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch runtime.GOOS {
		case "windows":
			cmd := exec.Command("systeminfo", "|", "findstr", "System Up Time")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		case "linux":
			cmd := exec.Command("uptime", "-p")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		case "darwin":
			cmd := exec.Command("uptime")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user",
	Long:  `Display current username.`,
	Run: func(cmd *cobra.Command, args []string) {
		user, _ := os.UserHomeDir()
		fmt.Printf("%s\n", user)
	},
}

var hostnameCmd = &cobra.Command{
	Use:   "hostname",
	Short: "Show hostname",
	Long:  `Display system hostname.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := os.Hostname()
		fmt.Printf("%s\n", name)
	},
}