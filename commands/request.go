package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

// formatJSON pretty-prints JSON with proper indentation for verbose output.
func formatJSON(body string, contentType string) string {
	if body == "" || contentType == "" || !strings.Contains(contentType, "application/json") {
		return body
	}
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(body), "", "  ")
	if err != nil {
		return body // Fallback to raw if unmarshalling fails
	}
	// Split and re-indent lines to align with "Body:"
	lines := strings.Split(prettyJSON.String(), "\n")
	for i, line := range lines {
		lines[i] = "    " + line // Add 4-space prefix to align with "Body:"
	}
	return strings.Join(lines, "\n")
}

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

func RequestCmd() *cobra.Command {
	var verbose bool
	var inferSchema bool

	cmd := &cobra.Command{
		Use:     "request <METHOD_name>",
		Aliases: []string{"r"},
		Short:   "Execute a request from a YAML file in the requests/ directory",
		Long: `Execute a request defined in a YAML file located in the requests/ directory.
The file should be named as METHOD_name.yaml (e.g., GET_config.yaml).
Use --infer-schema to generate a JTD schema from the response.
Use --verbose to show detailed request and response information.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fileName := args[0]
			parts := strings.SplitN(fileName, "_", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: invalid filename format '%s', expected 'METHOD_name' (e.g., 'POST_login')\n", fileName)
				os.Exit(1)
			}

			// Setup progress writer
			pw := progress.NewWriter()
			pw.SetAutoStop(false)
			pw.SetTrackerLength(25)
			pw.SetMessageWidth(24)
			pw.SetStyle(progress.StyleDefault)
			pw.SetUpdateFrequency(time.Millisecond * 100)
			pw.Style().Colors = progress.StyleColorsExample
			pw.Style().Visibility.Percentage = false
			pw.Style().Visibility.Value = false
			pw.Style().Visibility.TrackerOverall = false
			pw.Style().Visibility.Time = false

			go pw.Render()

			// Create tracker
			start := time.Now()
			tracker := &progress.Tracker{
				Message: fmt.Sprintf("%s idle", fileName),
				Total:   100,
			}
			pw.AppendTracker(tracker)

			// Update idle message periodically
			done := make(chan struct{})
			go func() {
				ticker := time.Tick(time.Millisecond * 100)
				for {
					select {
					case <-ticker:
						duration := time.Since(start)
						tracker.UpdateMessage(fmt.Sprintf("%s idle (%d ms)", fileName, duration.Milliseconds()))
					case <-done:
						return
					}
				}
			}()

			// Execute request
			resp, err := util.HandleRequest(fileName, inferSchema)
			close(done)
			if err != nil {
				pw.Stop()
				tracker.MarkAsErrored()
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Update tracker with colored status
			var statusColor text.Color
			switch {
			case resp.StatusCode >= 200 && resp.StatusCode < 300:
				statusColor = text.FgGreen
			case resp.StatusCode >= 400 && resp.StatusCode < 500:
				statusColor = text.FgYellow
			case resp.StatusCode >= 500:
				statusColor = text.FgRed
			default:
				statusColor = text.FgWhite
			}
			tracker.UpdateMessage(statusColor.Sprintf("%s %d", fileName, resp.StatusCode))
			tracker.MarkAsDone()

			// Wait for progress to finish
			for pw.IsRenderInProgress() {
				if pw.LengthActive() == 0 {
					pw.Stop()
				}
				time.Sleep(time.Millisecond * 100)
			}

			// Prepare bodies with JSON formatting
			reqContentType := ""
			respContentType := ""
			if len(resp.ReqHeaders["Content-Type"]) > 0 {
				reqContentType = strings.ToLower(strings.Split(resp.ReqHeaders["Content-Type"], ";")[0])
			}
			if len(resp.RespHeaders["Content-Type"]) > 0 {
				respContentType = strings.ToLower(strings.Split(resp.RespHeaders["Content-Type"][0], ";")[0])
			}
			reqBodyDisplay := formatJSON(resp.ReqBody, reqContentType)
			respBodyDisplay := formatJSON(resp.RespBody, respContentType)

			// Print response body
			fmt.Println("----------------")
			if verbose {
				fmt.Println(color.CyanString("Request:"))
				fmt.Println(color.HiBlueString("  Headers:"))
				if resp.ReqHeaders == nil || len(resp.ReqHeaders) == 0 {
					fmt.Printf(color.HiYellowString("    <Empty>\n"))
				}
				for k, v := range resp.ReqHeaders {
					fmt.Printf("    %s: %s\n", k, v)
				}
				fmt.Println(color.HiBlueString("  Body:"))
				if resp.ReqBody != "" {
					fmt.Printf("%s\n", reqBodyDisplay)
				} else {
					fmt.Printf(color.HiYellowString("    <Empty>\n"))
				}
				fmt.Println(color.CyanString("Response:"))
				fmt.Println(color.HiBlueString("  Headers:"))
				if resp.RespHeaders == nil || len(resp.RespHeaders) == 0 {
					fmt.Printf(color.HiYellowString("    <Empty>\n"))
				}
				for k, v := range resp.RespHeaders {
					for _, val := range v {
						fmt.Printf("    %s: %s\n", k, val)
					}
				}
				fmt.Println(color.HiBlueString("  Body:"))
			}
			if resp.RespBody != "" {
				fmt.Printf("%s\n", respBodyDisplay)
			} else {
				fmt.Printf(color.HiYellowString("    <Empty>\n"))
			}
		},
		ValidArgsFunction: requestCompletionFunc,
	}

	cmd.Flags().BoolVar(&inferSchema, "infer-schema", false, "Generate a JTD schema from the response")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed request and response information")

	return cmd
}
