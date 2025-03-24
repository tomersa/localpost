package commands

import (
	"fmt"
	"os"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a localpost project directory",
		Run: func(cmd *cobra.Command, args []string) {
			// Create LocalpostDir
			if err := os.MkdirAll(util.LocalpostDir, 0755); err != nil {
				fmt.Printf("Error creating %s directory: %v\n", util.LocalpostDir, err)
				os.Exit(1)
			}

			// Create RequestsDir
			if err := os.MkdirAll(util.RequestsDir, 0755); err != nil {
				fmt.Printf("Error creating %s directory: %v\n", util.RequestsDir, err)
				os.Exit(1)
			}

			// Create ResponsesDir // TODO: Will be implemented in future versions
			//if err := os.MkdirAll(util.ResponsesDir, 0755); err != nil {
			//	fmt.Printf("Error creating %s directory: %v\n", util.ResponsesDir, err)
			//	os.Exit(1)
			//}

			// Create ConfigFile
			config := struct {
				Env  string `yaml:"env"`
				Envs map[string]struct {
					Vars map[string]string `yaml:",inline"`
				} `yaml:"envs"`
			}{
				Env: "dev",
				Envs: map[string]struct {
					Vars map[string]string `yaml:",inline"`
				}{
					"dev": {Vars: make(map[string]string)},
				},
			}
			data, err := yaml.Marshal(&config)
			if err != nil {
				fmt.Printf("Error marshaling config: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(util.ConfigFilePath, data, 0644); err != nil {
				fmt.Printf("Error writing %s: %v\n", util.ConfigFilePath, err)
				os.Exit(1)
			}

			fmt.Printf("Initialized localpost project in %s\n", util.LocalpostDir)
		},
	}
}
