package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// webCmd wraps Playwright CLI through npx. Browser state is isolated by session.
var webCmd = &cobra.Command{Use: "web", Short: "Real-browser automation via Playwright CLI"}

func webAction(use, short string, min, max int, action string) *cobra.Command {
	command := &cobra.Command{Use: use, Short: short, Args: cobra.RangeArgs(min, max), RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaywright(cmd, append([]string{action}, args...)...)
	}}
	command.Flags().String("session", "default", "isolated Playwright session name")
	return command
}

var webOpenCmd = webAction("open <url>", "Open URL", 1, 1, "open")
var webSnapshotCmd = webAction("snapshot", "Capture accessible page snapshot before element-ref actions", 0, 0, "snapshot")
var webClickCmd = webAction("click <ref>", "Click snapshot element reference", 1, 1, "click")
var webFillCmd = webAction("fill <ref> <text>", "Fill snapshot element reference", 2, 2, "fill")
var webExtractCmd = webAction("extract <ref>", "Extract element text", 1, 1, "extract")
var webScreenshotCmd = webAction("screenshot [ref]", "Capture screenshot", 0, 1, "screenshot")
var webDownloadCmd = webAction("download <ref>", "Trigger download from element", 1, 1, "click")

func runPlaywright(cmd *cobra.Command, action ...string) error {
	npx, err := exec.LookPath("npx")
	if err != nil {
		return fmt.Errorf("npx not found; install Node.js/npm for web automation")
	}
	session, _ := cmd.Flags().GetString("session")
	if action[0] == "extract" {
		action = []string{"eval", "el => el.textContent", action[1]}
	}
	args := []string{"--yes", "--package", "@playwright/cli", "playwright-cli"}
	if session != "" {
		args = append(args, "--session", session)
	}
	args = append(args, action...)
	child := exec.Command(npx, args...)
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
	return child.Run()
}

func init() {
	webCmd.AddCommand(webOpenCmd, webSnapshotCmd, webClickCmd, webFillCmd, webExtractCmd, webScreenshotCmd, webDownloadCmd)
}
