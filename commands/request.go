package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func requestCompletionFunc(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var requestPaths []string
	err := filepath.Walk(util.RequestsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			relPath, err := filepath.Rel(util.RequestsDir, path)
			if err != nil {
				return err
			}
			relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
			requestPath := "/" + strings.TrimSuffix(relPath, ".yaml")
			if strings.HasPrefix(requestPath, "/"+strings.TrimPrefix(toComplete, "/")) {
				requestPaths = append(requestPaths, requestPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return requestPaths, cobra.ShellCompDirectiveNoSpace
}

func RequestCmd() *cobra.Command {
	var verbose bool
	var inferSchema bool

	cmd := &cobra.Command{
		Use:     "request <path>",
		Aliases: []string{"r"},
		Short:   "Execute a request from a YAML file in the requests/ directory",
		Long: `Execute a request defined in a YAML file located in the requests/ directory.
The path should be in the format /path/to/dir/METHOD (e.g., /user/POST or /api/v1/auth/login/POST).
Use --infer-schema to generate a JTD schema from the response.
Use --verbose to show detailed request and response information.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			requestPath := args[0]
			requestPath = strings.TrimPrefix(requestPath, "/")
			if requestPath == "" {
				fmt.Println("Error: request path cannot be empty")
				os.Exit(1)
			}

			// Validate that the path ends with a valid HTTP method
			parts := strings.Split(requestPath, "/")
			if len(parts) < 1 {
				fmt.Printf("Error: invalid request path '%s', expected format '/path/to/dir/METHOD'\n", requestPath)
				os.Exit(1)
			}
			method := parts[len(parts)-1]
			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true,
				"PATCH": true, "HEAD": true, "OPTIONS": true, "TRACE": true,
			}
			if !validMethods[method] {
				fmt.Printf("Error: invalid HTTP method '%s' in path, expected one of %v\n", method, validMethods)
				os.Exit(1)
			}

			filePath := filepath.Join(util.RequestsDir, requestPath+".yaml")
			_, err := util.HandleRequest(filePath, verbose, inferSchema)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
		ValidArgsFunction: requestCompletionFunc,
	}

	cmd.Flags().BoolVar(&inferSchema, "infer-schema", true, "Generate a JTD schema from the response")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed request and response information")

	return cmd
}
