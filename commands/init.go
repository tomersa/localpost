package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a localpost project directory",
		Run: func(cmd *cobra.Command, args []string) {
			// Create lpost directory
			if err := os.MkdirAll(util.LocalpostDir, 0755); err != nil {
				fmt.Printf("Error creating %s: %v\n", util.LocalpostDir, err)
				os.Exit(1)
			}

			// Create requests directory
			if err := os.MkdirAll(util.RequestsDir, 0755); err != nil {
				fmt.Printf("Error creating %s: %v\n", util.RequestsDir, err)
				os.Exit(1)
			}

			// Create config.yaml with login
			if _, err := os.Stat(util.ConfigFilePath); os.IsNotExist(err) {
				config := util.Config{
					Env: "dev",
					Envs: map[string]util.Env{
						"dev": {
							Vars: map[string]string{},
							Login: &util.LoginConfig{
								Request:     "POST_login.yaml",
								TriggeredBy: []int{401},
							},
						},
					},
				}
				data, err := yaml.Marshal(config)
				if err != nil {
					fmt.Printf("Error marshaling config: %v\n", err)
					os.Exit(1)
				}
				if err := os.WriteFile(util.ConfigFilePath, data, 0644); err != nil {
					fmt.Printf("Error writing %s: %v\n", util.ConfigFilePath, err)
					os.Exit(1)
				}
				fmt.Printf("Created %s with default environment and login\n", util.ConfigFilePath)
			} else {
				fmt.Printf("%s already exists\n", util.ConfigFilePath)
			}

			// Create default login request file
			if _, err := os.Stat(filepath.Join(util.RequestsDir, "auth/login/POST.yaml")); os.IsNotExist(err) {
				loginReq := util.RequestDefinition{
					Method: "POST",
					URL:    "{BASE_URL}/login",
					Body: util.Body{
						Json: map[string]interface{}{
							"username": "user",
							"password": "pass",
						},
					},
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				loginPath := filepath.Join(util.RequestsDir, "auth/login/POST.yaml")
				if err := os.MkdirAll(filepath.Dir(loginPath), 0755); err != nil {
					fmt.Printf("Error creating auth/login directory: %v\n", err)
					os.Exit(1)
				}
				loginData, err := yaml.Marshal(&loginReq)
				if err != nil {
					fmt.Printf("Error marshaling login request: %v\n", err)
					os.Exit(1)
				}
				if err := os.WriteFile(loginPath, loginData, 0644); err != nil {
					fmt.Printf("Error writing %s: %v\n", loginPath, err)
					os.Exit(1)
				}
				fmt.Printf("Created %s\n", loginPath)
			} else {
				fmt.Printf("%s already exists\n", filepath.Join(util.RequestsDir, "auth/login/POST.yaml"))
			}

			// Create .ephemeral.yaml
			if _, err := os.Stat(util.EphemeralFilePath); os.IsNotExist(err) {
				ephData := util.Ephemeral{
					Cookies: map[string]string{},
					Vars:    map[string]string{},
				}
				data, err := yaml.Marshal(ephData)
				if err != nil {
					fmt.Printf("Error marshaling ephemeral: %v\n", err)
					os.Exit(1)
				}
				if err := os.WriteFile(util.EphemeralFilePath, data, 0644); err != nil {
					fmt.Printf("Error writing %s: %v\n", util.EphemeralFilePath, err)
					os.Exit(1)
				}
				fmt.Printf("Created %s\n", util.EphemeralFilePath)
			} else {
				fmt.Printf("%s already exists\n", util.EphemeralFilePath)
			}

			// Create .gitignore
			if _, err := os.Stat(util.GitignoreFilePath); os.IsNotExist(err) {
				gitignoreContent := ".ephemeral.yaml\n"
				if err := os.WriteFile(util.GitignoreFilePath, []byte(gitignoreContent), 0644); err != nil {
					fmt.Printf("Error writing %s: %v\n", util.GitignoreFilePath, err)
					os.Exit(1)
				}
				fmt.Printf("Created %s\n", util.GitignoreFilePath)
			} else {
				fmt.Printf("%s already exists\n", util.GitignoreFilePath)
			}
		},
	}
}
