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

func NewAddRequestCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "add-request",
		Short:   "Create a new request YAML file interactively",
		Args:    cobra.NoArgs,
		GroupID: "requests",
		Run: func(cmd *cobra.Command, args []string) {
			var name, urlPath, method, contentType string
			var err error

			// Request Nickname (Name)
			prompt := promptui.Prompt{
				Label: "Enter request nickname (e.g., user-details)",
				Validate: func(input string) error {
					if strings.TrimSpace(input) == "" {
						return fmt.Errorf("nickname cannot be empty")
					}
					return nil
				},
			}
			name, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed: %v\n", err)
				os.Exit(1)
			}

			// URL Path
			prompt = promptui.Prompt{
				Label: "Enter URL (e.g., https://example.com/user/1 or {BASE_URL}/user/{id})",
				Validate: func(input string) error {
					if strings.TrimSpace(input) == "" {
						return fmt.Errorf("URL cannot be empty")
					}
					return nil
				},
			}
			urlPath, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed: %v\n", err)
				os.Exit(1)
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
				"POST":    true,
				"PUT":     true,
				"PATCH":   true,
				"OPTIONS": true,
				"TRACE":   true,
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
						body.Text = "example text content"
					}
					headers = map[string]string{"Content-Type": contentType}
				}
			}

			req := util.Request{
				Method:  method,
				Headers: headers,
				Body:    body,
				URL:     urlPath,
			}

			filePath := filepath.Join(util.DefaultRequestsDir, fmt.Sprintf("%s_%s.yaml", method, name))
			if err := os.MkdirAll(util.DefaultRequestsDir, 0755); err != nil {
				fmt.Printf("Error creating requests directory: %v\n", err)
				os.Exit(1)
			}
			data, err := yaml.Marshal(&req)
			if err != nil {
				fmt.Printf("Error marshaling request: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(filePath, data, 0644); err != nil {
				fmt.Printf("Error writing request file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created new request file: %s\n", filePath)
		},
	}
}
