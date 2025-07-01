package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func AddRequestCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add-request",
		Short:   "Create a new request YAML file interactively with hierarchical storage",
		Args:    cobra.NoArgs,
		GroupID: "requests",
		Run: func(cmd *cobra.Command, args []string) {
			var urlPath, method, contentType string
			var err error

			// URL Path
			prompt := promptui.Prompt{
				Label: "Enter URL path (e.g., /user or /api/v1/auth/login, without protocol or domain)",
				Validate: func(input string) error {
					input = strings.TrimSpace(input)
					if input == "" {
						return fmt.Errorf("URL path cannot be empty")
					}
					// Reject protocol or domain (e.g., http://, https://, www.google.com)
					if strings.Contains(input, "://") || strings.HasPrefix(input, "www.") {
						return fmt.Errorf("enter only the path (e.g., /user), not a full URL")
					}
					return nil
				},
				Stdout: os.Stdout,
			}
			urlPath, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed: %v\n", err)
				os.Exit(1)
			}

			// Derive directory path from URL
			// Remove leading/trailing slashes, query params, and fragments
			pathPart := strings.TrimPrefix(urlPath, "/")
			pathPart = strings.Split(pathPart, "?")[0]
			pathPart = strings.Split(pathPart, "#")[0]
			dirPath := pathPart
			if dirPath == "" {
				dirPath = "root" // Fallback for root URLs
			}

			// HTTP Method Menu
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE"}
			selectPrompt := promptui.Select{
				Label: "Select HTTP method",
				Items: methods,
			}
			_, method, err = selectPrompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed: %v\n", err)
				os.Exit(1)
			}

			// Body Type Menu (only for body-supporting methods)
			body := util.Body{}
			bodyMethods := map[string]bool{
				"POST":  true,
				"PUT":   true,
				"PATCH": true,
			}
			var headers map[string]string
			if bodyMethods[method] {
				contentTypes := []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data", "text/plain", "none"}
				selectPrompt = promptui.Select{
					Label: "Select body type (sets Content-Type header)",
					Items: contentTypes,
				}
				_, contentType, err = selectPrompt.Run()
				if err != nil {
					fmt.Printf("Prompt failed: %v\n", err)
					os.Exit(1)
				}

				if contentType != "none" {
					switch contentType {
					case "application/json":
						body.Json = map[string]interface{}{"key": "value"}
					case "application/x-www-form-urlencoded":
						body.FormUrlEncoded = map[string]string{"key": "value"}
					case "multipart/form-data":
						body.Form.Fields = map[string]string{"field": "value"}
						body.Form.Files = map[string]string{"file": "/path/to/file"}
					case "text/plain":
						body.Text = "example text"
					}
					headers = map[string]string{"Content-Type": contentType}
				}
			}

			req := util.RequestDefinition{
				Headers: headers,
				Body:    body,
			}

			// Create file path: requests/<dirPath>/<method>.yaml
			fileName := fmt.Sprintf("%s.yaml", method)
			filePath := filepath.Join(util.RequestsDir, dirPath, fileName)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				fmt.Printf("Error creating request directory: %v\n", err)
				os.Exit(1)
			}

			data, err := yaml.Marshal(&req)
			if err != nil {
				fmt.Printf("Error marshaling request: %v\n", err)
				os.Exit(1)
			}
			yamlStr := string(data)
			yamlStr = strings.ReplaceAll(yamlStr, "key: value", "#key: value")
			yamlStr = strings.ReplaceAll(yamlStr, "field: value", "#field: value")
			yamlStr = strings.ReplaceAll(yamlStr, "example text", "#example text")

			if err := os.WriteFile(filePath, []byte(yamlStr), 0644); err != nil {
				fmt.Printf("Error writing request file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created new request file: %s\n", filePath)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}
