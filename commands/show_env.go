package commands

import (
	"fmt"
	"os"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func NewShowEnvCommand() *cobra.Command {
	var showAll bool
	cmd := &cobra.Command{
		Use:     "show-env",
		Short:   "Show the current environment variables",
		Args:    cobra.NoArgs,
		GroupID: "environment",
		Run: func(cmd *cobra.Command, args []string) {
			if showAll {
				configFile, err := util.GetConfig()
				if err != nil {
					fmt.Printf("Error loading .localpost: %v\n", err)
					os.Exit(1)
				}

				fmt.Println(configFile)
				return
			}

			env, err := util.LoadEnv()
			if err != nil {
				fmt.Printf("Error loading env: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Current environment: %s\n", env.Name)
			if len(env.Vars) == 0 {
				fmt.Println("No variables set")
			} else {
				for key, value := range env.Vars {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
		},
	}
	cmd.Flags().BoolVar(&showAll, "all", false, "Show the entire .localpost configuration")
	return cmd
}
