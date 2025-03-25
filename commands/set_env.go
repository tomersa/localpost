package commands

import (
	"fmt"
	"os"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func NewSetEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "set-env <env>",
		Short:   "Set the current environment",
		Args:    cobra.ExactArgs(1),
		GroupID: "environment",
		Run: func(cmd *cobra.Command, args []string) {
			err := util.SetEnv(args[0])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Environment set to %s\n", args[0])
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}
