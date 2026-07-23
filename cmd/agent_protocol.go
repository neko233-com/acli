package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const agentProtocolVersion = "v1"

var schemaCmd = &cobra.Command{
	Use: "schema [command...]", Short: "Emit JSON Schema-like command contract for agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		target, remaining, err := rootCmd.Find(args)
		if err != nil {
			return err
		}
		if len(remaining) > 0 {
			return fmt.Errorf("unknown command path: %s", strings.Join(args, " "))
		}
		return json.NewEncoder(os.Stdout).Encode(commandSchema(target))
	},
}

var agentProtocolCmd = &cobra.Command{Use: "protocol", Short: "Print universal agent execution protocol", RunE: func(cmd *cobra.Command, args []string) error {
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"version": agentProtocolVersion, "discover": "acli schema <command...>", "errors": "stderr text; exit 0 success, non-zero failure", "safety": []string{"--dry-run previews supported mutations", "--yes confirms destructive actions", "prefer --out over in-place mutation"}, "formats": []string{"--json", "--jsonl"}})
}}

func commandSchema(command *cobra.Command) map[string]any {
	flags := []map[string]any{}
	command.InheritedFlags().VisitAll(func(flag *pflag.Flag) { flags = append(flags, schemaFlag(flag)) })
	command.LocalFlags().VisitAll(func(flag *pflag.Flag) { flags = append(flags, schemaFlag(flag)) })
	return map[string]any{"protocol": agentProtocolVersion, "command": command.CommandPath(), "use": command.Use, "summary": command.Short, "args": command.Args != nil, "flags": flags}
}
func schemaFlag(flag *pflag.Flag) map[string]any {
	return map[string]any{"name": flag.Name, "type": flag.Value.Type(), "default": flag.DefValue, "description": flag.Usage}
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "request JSON output when command supports it")
	rootCmd.PersistentFlags().Bool("jsonl", false, "request JSON Lines output when command supports it")
	rootCmd.PersistentFlags().Bool("dry-run", false, "preview supported mutation without changing state")
	rootCmd.PersistentFlags().Bool("yes", false, "confirm supported destructive action")
	rootCmd.PersistentFlags().Bool("stdin", false, "read supported payload from stdin")
	rootCmd.PersistentFlags().String("out", "", "write supported result to output path")
	rootCmd.PersistentFlags().Int("page", 1, "page number for supported list commands")
	rootCmd.PersistentFlags().Int("page-size", 100, "page size for supported list commands")
	rootCmd.PersistentFlags().String("filter", "", "filter expression for supported list commands")
	rootCmd.AddCommand(schemaCmd)
	agentCmd.AddCommand(agentProtocolCmd)
}
