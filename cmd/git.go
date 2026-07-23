package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{Use: "git", Short: "Git workflows through one agent-safe CLI"}
var gitStatusCmd = gitCommand("status", "Show worktree status", "status", "--short", "--branch")
var gitDiffCmd = gitCommand("diff [args...]", "Show diff", "diff")
var gitLogCmd = gitCommand("log [args...]", "Show concise history", "log", "--oneline", "--decorate", "-20")
var gitBranchCmd = gitCommand("branch [args...]", "List or create branches", "branch")
var gitFetchCmd = gitCommand("fetch [args...]", "Fetch remotes", "fetch", "--prune")
var gitPullCmd = gitCommand("pull [args...]", "Pull current branch", "pull", "--ff-only")
var gitPushCmd = gitDangerCommand("push [args...]", "Push refs", "push")
var gitMergeCmd = gitDangerCommand("merge <branch>", "Merge branch", "merge")
var gitRebaseCmd = gitDangerCommand("rebase <branch>", "Rebase current branch", "rebase")
var gitStashCmd = gitCommand("stash [args...]", "Stash worktree changes", "stash")
var gitPRCmd = &cobra.Command{Use: "pr [args...]", Short: "Create pull request via gh", RunE: func(cmd *cobra.Command, args []string) error {
	return runDangerTool(cmd, "gh", append([]string{"pr", "create"}, args...)...)
}}

var gitCommitCmd = &cobra.Command{Use: "commit -m <message>", Short: "Commit staged changes", RunE: func(cmd *cobra.Command, args []string) error {
	if err := requireMutationApproval(cmd, "git commit"); err != nil {
		return err
	}
	message, _ := cmd.Flags().GetString("message")
	if message == "" {
		return fmt.Errorf("-m/--message is required")
	}
	return runGit(cmd, "commit", "-m", message)
}}
var gitAddCmd = &cobra.Command{Use: "add <path...>", Short: "Stage explicit paths", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	return runGit(cmd, append([]string{"add", "--"}, args...)...)
}}

func gitCommand(use, short string, base ...string) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, RunE: func(cmd *cobra.Command, args []string) error { return runGit(cmd, append(base, args...)...) }}
}
func gitDangerCommand(use, short string, base ...string) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireMutationApproval(cmd, "git "+base[0]); err != nil {
			return err
		}
		return runGit(cmd, append(base, args...)...)
	}}
}
func requireMutationApproval(cmd *cobra.Command, action string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "dry_run": true, "action": action})
	}
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		return fmt.Errorf("%s changes state; pass --yes or --dry-run", action)
	}
	return nil
}
func runDangerTool(cmd *cobra.Command, tool string, args ...string) error {
	if err := requireMutationApproval(cmd, tool+" "+strings.Join(args[:2], " ")); err != nil {
		return err
	}
	path, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("%s not found in PATH", tool)
	}
	child := exec.Command(path, args...)
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
	return child.Run()
}
func runGit(cmd *cobra.Command, args ...string) error {
	path, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH")
	}
	child := exec.Command(path, args...)
	child.Dir, _ = cmd.Flags().GetString("dir")
	if child.Dir == "" {
		child.Dir = "."
	}
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
	return child.Run()
}

func init() {
	commands := []*cobra.Command{gitStatusCmd, gitDiffCmd, gitLogCmd, gitBranchCmd, gitFetchCmd, gitPullCmd, gitPushCmd, gitMergeCmd, gitRebaseCmd, gitStashCmd, gitPRCmd, gitAddCmd, gitCommitCmd}
	for _, command := range commands {
		command.Flags().String("dir", ".", "repository directory")
	}
	gitCommitCmd.Flags().StringP("message", "m", "", "commit message")
	gitCmd.AddCommand(commands...)
}
