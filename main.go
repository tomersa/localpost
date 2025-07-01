package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/moshe5745/localpost/commands"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

var (
	version = "0.0.1"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "localpost",
		Version: Version,
		Short:   "A CLI tool to manage and execute HTTP requests",
		Long:    `A tool to save and execute HTTP requests stored in a Git repository.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip for init and completion commands
			if cmd.Name() == "init" || cmd.Name() == "completion" {
				return nil
			}

			// Check localpost context for all other commands
			if err := util.CheckRepoContext(); err != nil {
				red := color.New(color.FgRed).SprintFunc()
				cmd.SilenceUsage = true
				return fmt.Errorf("%s\nMake sure you running in the right directory or run 'lpost init' to init localpost.", red(err))
			}

			// Handle --env flag
			flag := cmd.PersistentFlags().Lookup("env")
			if flag != nil {
				if flagEnv := flag.Value.String(); flagEnv != "" {
					if err := util.SetEnv(flagEnv); err != nil {
						return fmt.Errorf("error setting environment: %v", err)
					}
				}
			}
			return nil
		},
	}
	rootCmd.PersistentFlags().StringP("env", "e", "", "Environment to use (e.g., dev, prod); defaults to .localpost-config or 'dev'")

	rootCmd.AddGroup(&cobra.Group{ID: "requests", Title: "RequestDefinition Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "environment", Title: "Environment Commands"})

	rootCmd.AddCommand(commands.InitCmd())
	rootCmd.AddCommand(commands.AddRequestCmd())
	rootCmd.AddCommand(commands.RequestCmd())
	rootCmd.AddCommand(commands.TestCmd())
	rootCmd.AddCommand(commands.SetEnvCmd())
	rootCmd.AddCommand(commands.SetEnvVarCmd())
	rootCmd.AddCommand(commands.ShowEnvCmd())
	rootCmd.AddCommand(commands.CompletionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
