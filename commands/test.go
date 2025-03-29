package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jsontypedef/json-typedef-go"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func removeExtension(filename string) string {
	base := filepath.Base(filename)  // Gets the filename without path
	ext := filepath.Ext(base)        // Gets the extension (e.g., ".yaml")
	return base[:len(base)-len(ext)] // Returns filename without extension
}

func NewTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run all requests and validate against stored JTD schemas",
		Run: func(cmd *cobra.Command, args []string) {
			files, err := os.ReadDir(util.RequestsDir)
			if err != nil {
				fmt.Printf("Error reading requests dir: %v\n", err)
				os.Exit(1)
			}

			for _, file := range files {
				fileName := removeExtension(file.Name())
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
					resp, err := util.HandleRequest(fileName)
					if resp.StatusCode >= 400 {
						fmt.Printf("Request failed skiping validation: %s\n", file.Name())
						continue
					}
					if err != nil {
						fmt.Printf("Error executing %s: %v\n", file.Name(), err)
						continue
					}
					schemaPath := filepath.Join(util.SchemasDir, fileName+".jtd.json")
					schemaData, err := os.ReadFile(schemaPath)
					if err != nil {
						fmt.Printf("No schema for %s: %v\n", file.Name(), err)
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
