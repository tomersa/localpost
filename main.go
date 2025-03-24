package main

import (
	"fmt"
	"github.com/fatih/color"
	"os"

	"github.com/moshe5745/localpost/commands"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "localpost",
		Short: "A CLI tool to manage and execute HTTP requests",
		Long:  `A tool to save and execute HTTP requests stored in a Git repository.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip for init and completion commands
			if cmd.Name() == "init" || cmd.Name() == "completion" {
				return nil
			}

			// Check localpost/ context for all other commands
			if err := util.CheckRepoContext(); err != nil {
				red := color.New(color.FgRed).SprintFunc()
				return fmt.Errorf("%s", red(err))
			}
			// Handle --env flag
			flag := cmd.PersistentFlags().Lookup("env")
			if flag != nil {
				if flagEnv := flag.Value.String(); flagEnv != "" {
					_, err := util.SetEnvName(flagEnv)
					if err != nil {
						return fmt.Errorf("error setting environment: %v", err)
					}
				}
			}
			return nil
		},
	}
	rootCmd.PersistentFlags().StringP("env", "e", "", "Environment to use (e.g., dev, prod); defaults to .localpost-config or 'dev'")

	// Define command groups
	rootCmd.AddGroup(&cobra.Group{ID: "requests", Title: "Request Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "environment", Title: "Environment Commands"})

	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewAddRequestCommand())
	rootCmd.AddCommand(commands.NewRequestCommand())
	rootCmd.AddCommand(commands.NewSetEnvCommand())
	rootCmd.AddCommand(commands.NewSetEnvVarCommand())
	rootCmd.AddCommand(commands.NewShowEnvCommand())
	rootCmd.AddCommand(commands.NewCompletionCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
