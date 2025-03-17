package commands

import (
	"fmt"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
	"os"
)

func NewSetEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-env [env]",
		Short: "Set the current environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			env, err := util.SetEnvName(args[0])
			if err != nil {
				fmt.Printf("Error setting environment: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Set current environment to '%s'\n", env.Name)
		},
	}
}
