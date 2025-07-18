package commands

import (
	"fmt"
	"os"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func SetEnvVarCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set-env-var <key> <value>",
		Short:   "Set an environment variable for the current environment",
		Args:    cobra.ExactArgs(2),
		GroupID: "environment",
		Run: func(cmd *cobra.Command, args []string) {
			// Set the environment variable using updateConfig
			if err := util.SetEnvVar(args[0], args[1]); err != nil {
				fmt.Printf("Error setting %s: %v\n", args[0], err)
				os.Exit(1)
			}

			// Load the updated environment to get the name
			updatedEnv, err := util.LoadEnv()
			if err != nil {
				fmt.Printf("Error loading updated environment: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Set %s to %s for environment '%s'\n", args[0], args[1], updatedEnv.Name)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}
