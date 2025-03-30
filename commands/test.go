package commands

import (
	"encoding/json"
	"fmt"
	jtd "github.com/jsontypedef/json-typedef-go"
	"os"
	"path/filepath"
	"strings"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func removeExtension(filename string) string {
	base := filepath.Base(filename)  // Gets the filename without path
	ext := filepath.Ext(base)        // Gets the extension (e.g., ".yaml")
	return base[:len(base)-len(ext)] // Returns filename without extension
}

func TestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run all requests and validate against stored JTD schemas",
		Run: func(cmd *cobra.Command, args []string) {
			if err := util.ClearCookies(); err != nil {
				fmt.Printf("Error clearing cookies: %v\n", err)
				os.Exit(1)
			}

			env, err := util.LoadEnv()
			if err != nil {
				fmt.Printf("Error loading env: %v\n", err)
				os.Exit(1)
			}
			if env.Login == nil || env.Login.Request == "" {
				fmt.Println("No login request defined in config.yaml")
			} else {
				_, err := util.HandleRequest(env.Login.Request) // No schema inference here
				if err != nil {
					fmt.Printf("Error executing login request %s: %v\n", env.Login.Request, err)
					os.Exit(1)
				}
			}

			files, err := os.ReadDir(util.RequestsDir)
			if err != nil {
				fmt.Printf("Error reading requests dir: %v\n", err)
				os.Exit(1)
			}

			for _, file := range files {
				fileName := removeExtension(file.Name())
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
					if env.Login != nil && fileName == env.Login.Request {
						continue
					}

					resp, err := util.HandleRequest(fileName)
					if err != nil {
						fmt.Printf("Error executing %s: %v\n", file.Name(), err)
						continue
					}

					schemaPath := filepath.Join(util.SchemasDir, fileName+".jtd.json")
					schemaData, err := os.ReadFile(schemaPath)
					if err != nil {
						fmt.Printf("%s: no schema found - %v\n", fileName, err)
						continue
					}
					var schema jtd.Schema
					if err := json.Unmarshal(schemaData, &schema); err != nil {
						fmt.Printf("Error parsing schema %s: %v\n", schemaPath, err)
						continue
					}
					var doc interface{}
					if err := json.Unmarshal([]byte(resp.RespBody), &doc); err != nil {
						fmt.Printf("Error parsing response %s: %v\n", file.Name(), err)
						continue
					}
					if validateErr, _ := jtd.Validate(schema, doc); err != nil {
						fmt.Printf("Validation failed for %s: %v\n", file.Name(), validateErr)
					} else {
						fmt.Printf("%s: schema valid\n", file.Name())
					}
				}
			}
		},
	}
}
