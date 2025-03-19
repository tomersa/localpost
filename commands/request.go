package commands

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func requestCompletionFunc(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	files, err := os.ReadDir(util.RequestsDir)
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
		Aliases: []string{"-r"},
		Short:   "Execute an HTTP request from a YAML file",
		Args:    cobra.ExactArgs(1),
		GroupID: "requests",
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			parts := strings.SplitN(input, "_", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: invalid input format '%s', expected 'METHOD_name' (e.g., 'POST_login')\n", input)
				os.Exit(1)
			}
			method := parts[0]
			if method != strings.ToUpper(method) {
				fmt.Printf("Error: method '%s' must be uppercase (e.g., GET, POST)\n", method)
				os.Exit(1)
			}

			filePath := filepath.Join(util.RequestsDir, fmt.Sprintf("%s.yaml", input))
			req, err := util.ParseRequest(filePath)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			req.Method = method

			resp, err := util.ExecuteRequest(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			var statusColor func(a ...interface{}) string
			switch {
			case strings.HasPrefix(resp.Status, "2"):
				statusColor = color.New(color.FgGreen).SprintFunc()
			case strings.HasPrefix(resp.Status, "4"):
				statusColor = color.New(color.FgYellow).SprintFunc()
			case strings.HasPrefix(resp.Status, "5"):
				statusColor = color.New(color.FgRed).SprintFunc()
			default:
				statusColor = color.New(color.FgWhite).SprintFunc()
			}

			// Regular output: table
			if !verbose {
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"STATUS", "TIME"})
				t.AppendRow(table.Row{
					statusColor(resp.Status),
					fmt.Sprintf("%dms", resp.Duration.Milliseconds()),
				})
				if !verbose && resp.RespBody != "" {
					t.AppendSeparator()
					t.AppendRow(table.Row{"BODY", "BODY"}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft})
					t.AppendSeparator()
					t.AppendRow(table.Row{resp.RespBody, resp.RespBody}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft})
				}
				t.Render()

				return
			}

			fmt.Println("-----")

			fmt.Printf("Status: %s\n", statusColor(resp.Status))
			fmt.Printf("Time: %dms\n", resp.Duration.Milliseconds())
			fmt.Printf("URL: %s\n", resp.ReqURL)

			fmt.Println("-----")

			fmt.Println(color.CyanString("Request"))
			fmt.Println(color.HiBlueString("  Headers:"))
			for k, v := range resp.ReqHeaders {
				fmt.Printf("    %s: %s\n", k, v)
			}
			fmt.Println(color.CyanString("Request"))
			fmt.Println(color.HiBlueString("  Body:"))
			fmt.Printf("    %s\n", resp.ReqBody)

			fmt.Println("-----")

			fmt.Println(color.CyanString("Response"))
			fmt.Println(color.HiBlueString("  Headers:"))
			for k, v := range resp.RespHeaders {
				fmt.Printf("    %s: %s\n", k, v)
			}
			fmt.Println(color.CyanString("Request"))
			fmt.Println(color.HiBlueString("  Body:"))
			fmt.Printf("    %s\n", resp.RespBody)
		},
		ValidArgsFunction: requestCompletionFunc,
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full request and response details")
	return cmd
}
