package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/neko233-com/acme233/pkg/acmego"
	"github.com/spf13/cobra"
)

// acmeCmd wraps neko233-com/acme-go's public acmego package. Its module path is
// github.com/neko233-com/acme233; that is the path published by upstream.
var acmeCmd = &cobra.Command{Use: "acme", Short: "ACME certificate lifecycle (neko233-com/acme-go)"}

func acmeConfig(command *cobra.Command) (*acmego.Config, error) {
	path, _ := command.Flags().GetString("config")
	return acmego.Load(path)
}
func acmeName(command *cobra.Command) string {
	name, _ := command.Flags().GetString("name")
	return name
}

var acmeValidateCmd = &cobra.Command{Use: "validate", Short: "Validate ACME config", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	return acmego.Validate(cfg, os.Stdout)
}}
var acmePlanCmd = &cobra.Command{Use: "plan", Short: "Preview issue/renew actions", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	return acmego.Plan(cfg, acmeName(cmd), os.Stdout)
}}
var acmePathsCmd = &cobra.Command{Use: "paths", Short: "Print certificate paths", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	return acmego.Paths(cfg, acmeName(cmd), os.Stdout)
}}
var acmeListCmd = &cobra.Command{Use: "list", Short: "List configured certificates", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	return acmego.List(cfg, os.Stdout)
}}
var acmeProvidersCmd = &cobra.Command{Use: "providers", Short: "List supported DNS providers", RunE: func(cmd *cobra.Command, args []string) error { return acmego.Providers(os.Stdout) }}

var acmeInfoCmd = &cobra.Command{Use: "info", Short: "Read certificate state", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	if acmeName(cmd) == "" {
		return fmt.Errorf("--name is required")
	}
	return acmego.Info(cfg, acmeName(cmd), os.Stdout)
}}
var acmeIssueCmd = &cobra.Command{Use: "issue", Short: "Issue certificate", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	force, _ := cmd.Flags().GetBool("force")
	_, err = acmego.Issue(cfg, acmeName(cmd), force, os.Stdout)
	return err
}}
var acmeRenewCmd = &cobra.Command{Use: "renew", Short: "Renew due certificate", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	force, _ := cmd.Flags().GetBool("force")
	_, err = acmego.Renew(cfg, acmeName(cmd), force, os.Stdout)
	return err
}}
var acmeRevokeCmd = &cobra.Command{Use: "revoke", Short: "Revoke certificate", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	if acmeName(cmd) == "" {
		return fmt.Errorf("--name is required")
	}
	return acmego.Revoke(cfg, acmeName(cmd), os.Stdout)
}}
var acmeInstallCmd = &cobra.Command{Use: "install", Short: "Copy certificate to configured destinations", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	if acmeName(cmd) == "" {
		return fmt.Errorf("--name is required")
	}
	return acmego.Install(cfg, acmeName(cmd), os.Stdout)
}}
var acmeDeployCmd = &cobra.Command{Use: "deploy", Short: "Run configured certificate deployment", RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := acmeConfig(cmd)
	if err != nil {
		return err
	}
	if acmeName(cmd) == "" {
		return fmt.Errorf("--name is required")
	}
	return acmego.Deploy(cfg, acmeName(cmd), os.Stdout)
}}

var acmeWatchCmd = &cobra.Command{
	Use: "watch", Aliases: []string{"auto-renew"}, Short: "Run renewal loop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := acmeConfig(cmd)
		if err != nil {
			return err
		}
		interval, _ := cmd.Flags().GetDuration("interval")
		force, _ := cmd.Flags().GetBool("force")
		once, _ := cmd.Flags().GetBool("once")
		if once {
			_, err := acmego.Renew(cfg, acmeName(cmd), force, os.Stdout)
			return err
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		return acmego.AutoRenewLoop(ctx, cfg, acmego.AutoRenewOptions{Name: acmeName(cmd), Force: force, Interval: interval, RunImmediately: true, Out: os.Stdout})
	},
}

func init() {
	all := []*cobra.Command{acmeValidateCmd, acmePlanCmd, acmePathsCmd, acmeListCmd, acmeInfoCmd, acmeIssueCmd, acmeRenewCmd, acmeRevokeCmd, acmeInstallCmd, acmeDeployCmd, acmeWatchCmd}
	for _, command := range all {
		command.Flags().String("config", "config_acme.json", "ACME config path")
	}
	for _, command := range []*cobra.Command{acmePlanCmd, acmePathsCmd, acmeInfoCmd, acmeIssueCmd, acmeRenewCmd, acmeRevokeCmd, acmeInstallCmd, acmeDeployCmd, acmeWatchCmd} {
		command.Flags().String("name", "", "certificate entry name")
	}
	for _, command := range []*cobra.Command{acmeIssueCmd, acmeRenewCmd, acmeWatchCmd} {
		command.Flags().Bool("force", false, "force operation")
	}
	acmeWatchCmd.Flags().Duration("interval", 24*time.Hour, "renew interval")
	acmeWatchCmd.Flags().Bool("once", false, "renew once then exit")
	acmeCmd.AddCommand(acmeValidateCmd, acmePlanCmd, acmePathsCmd, acmeListCmd, acmeInfoCmd, acmeProvidersCmd, acmeIssueCmd, acmeRenewCmd, acmeRevokeCmd, acmeInstallCmd, acmeDeployCmd, acmeWatchCmd)
}
