package main

import (
	"fmt"
	"github.com/moshe5745/localpost/util"
	"os"

	"github.com/moshe5745/localpost/commands" // Adjust to your actual package path
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "localpost",
		Short: "A CLI tool to manage and execute HTTP requests",
		Long:  `A tool to save and execute HTTP requests stored in a Git repository.`,
	}
	rootCmd.PersistentFlags().StringP("env", "e", "", "Environment to use (e.g., dev, prod); defaults to .localpost-config or 'dev'")

	// PersistentPreRun with safe flag access and completion skip
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Skip for completion command
		if cmd.Name() == "completion" {
			return
		}

		flag := cmd.PersistentFlags().Lookup("env")
		if flag != nil {
			if flagEnv := flag.Value.String(); flagEnv != "" {
				_, err := util.SetEnvName(flagEnv)
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintf("Error setting environment: %v\n", err))
					os.Exit(1)
				}
			}
		}
	}

	// Define command groups
	rootCmd.AddGroup(&cobra.Group{ID: "requests", Title: "Request Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "environment", Title: "Environment Commands"})

	// Add commands with their respective groups
	rootCmd.AddCommand(commands.NewRequestCommand())
	rootCmd.AddCommand(commands.NewAddRequestCommand())
	rootCmd.AddCommand(commands.NewSetEnvCommand())
	rootCmd.AddCommand(commands.NewShowEnvCommand())
	rootCmd.AddCommand(commands.NewSetEnvVarCommand())
	rootCmd.AddCommand(commands.NewCompletionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
