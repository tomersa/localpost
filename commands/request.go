package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func requestCompletionFunc(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	files, err := os.ReadDir(util.DefaultRequestsDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var requestFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			fullName := strings.TrimSuffix(file.Name(), ".yaml")
			if strings.HasPrefix(fullName, toComplete) {
				requestFiles = append(requestFiles, fullName)
			}
		}
	}

	if len(requestFiles) > 0 {
		return requestFiles, cobra.ShellCompDirectiveNoSpace
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func NewRequestCommand() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:     "request <METHOD_name>",
		Aliases: []string{"-r"}, // Add -r shorthand
		Short:   "Execute an HTTP request from a YAML file",
		Args:    cobra.ExactArgs(1),
		GroupID: "requests",
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			parts := strings.SplitN(input, "_", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: invalid input format '%s', expected 'METHOD_name' (e.g., 'POST_login-form')\n", input)
				os.Exit(1)
			}
			method := parts[0]
			if method != strings.ToUpper(method) {
				fmt.Printf("Error: method '%s' must be uppercase (e.g., GET, POST)\n", method)
				os.Exit(1)
			}

			filePath := filepath.Join(util.DefaultRequestsDir, fmt.Sprintf("%s.yaml", input))
			req, err := util.ParseRequest(filePath)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			req.Method = method

			reqHeaders, reqBody, status, respHeaders, respBody, duration, err := util.ExecuteRequest(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			var statusColor func(a ...interface{}) string
			switch {
			case strings.HasPrefix(status, "2"):
				statusColor = color.New(color.FgGreen).SprintFunc()
			case strings.HasPrefix(status, "4"):
				statusColor = color.New(color.FgYellow).SprintFunc()
			case strings.HasPrefix(status, "5"):
				statusColor = color.New(color.FgRed).SprintFunc()
			default:
				statusColor = color.New(color.FgWhite).SprintFunc()
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Status", "Time"})
			t.AppendRow(table.Row{
				statusColor(status),
				fmt.Sprintf("%dms", duration.Milliseconds()),
			})

			if !verbose && respBody != "" {
				t.AppendSeparator()
				t.AppendRow(table.Row{"BODY", "BODY"}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft})
				t.AppendSeparator()
				t.AppendRow(table.Row{respBody, respBody}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft})
			}
			t.Render()

			if verbose {
				fmt.Println("\nVerbose Details:")
				fmt.Printf("URL: %s\n", req.URL)
				fmt.Println(color.CyanString("Request Headers:"))
				for k, v := range reqHeaders {
					fmt.Printf("  %s: %s\n", k, v)
				}
				fmt.Println(color.CyanString("Request Body:"))
				fmt.Printf("  %s\n", reqBody)
				fmt.Println(color.CyanString("Response Headers:"))
				for k, v := range respHeaders {
					fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
				}
				fmt.Println(color.CyanString("Response Body:"))
				fmt.Printf("  %s\n", respBody)
			}
		},
		ValidArgsFunction: requestCompletionFunc,
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full request and response details")
	return cmd
}
