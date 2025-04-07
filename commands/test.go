package commands

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/text"
	jtd "github.com/jsontypedef/json-typedef-go"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func TestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run all requests and validate against stored JTD schemas",
		Run: func(cmd *cobra.Command, args []string) {
			// Clear cookies
			if err := util.ClearCookies(); err != nil {
				fmt.Printf("Error clearing cookies: %v\n", err)
				os.Exit(1)
			}

			// Execute login request with response output
			env, err := util.LoadEnv()
			if err != nil {
				fmt.Printf("Error loading env: %v\n", err)
				os.Exit(1)
			}
			if env.Login != nil && env.Login.Request != "" {
				_, err := util.HandleRequest(env.Login.Request, false)
				if err != nil {
					fmt.Printf("Error executing login request %s: %v\n", env.Login.Request, err)
					os.Exit(1)
				}
			}

			// Collect requests
			files, err := os.ReadDir(util.RequestsDir)
			if err != nil {
				fmt.Printf("Error reading requests dir: %v\n", err)
				os.Exit(1)
			}

			// Setup progress writer
			pw := progress.NewWriter()
			pw.SetAutoStop(false)
			pw.SetTrackerLength(25)
			pw.SetMessageWidth(40)
			pw.SetStyle(progress.StyleDefault)
			pw.SetUpdateFrequency(time.Millisecond * 100)
			pw.SetTrackerPosition(progress.PositionRight)
			pw.Style().Colors = progress.StyleColorsExample
			pw.Style().Visibility.Percentage = false
			pw.Style().Visibility.Value = false
			pw.Style().Visibility.TrackerOverall = false
			pw.Style().Visibility.Time = true

			go pw.Render()

			// Track failures and requests
			var wg sync.WaitGroup
			failed := false
			mu := sync.Mutex{}
			trackers := make(map[string]*progress.Tracker)

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
					fileName := strings.TrimSuffix(file.Name(), ".yaml")
					if env.Login != nil && fileName == strings.TrimSuffix(env.Login.Request, ".yaml") {
						continue // Skip login request
					}

					wg.Add(1)
					tracker := &progress.Tracker{
						Message: fmt.Sprintf("%s idle", fileName),
						Total:   0,
					}
					pw.AppendTracker(tracker)
					trackers[fileName] = tracker

					go func(fn string, t *progress.Tracker) {
						defer wg.Done()

						// Execute request
						resp, err := util.HandleRequest(fn, false)
						if err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.UpdateMessage(fmt.Sprintf("%s failed: %v", fn, err))
							t.MarkAsErrored()
							return
						}

						// Validate schema
						schemaPath := filepath.Join(util.SchemasDir, fn+".jtd.json")
						schemaData, err := os.ReadFile(schemaPath)
						if err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.UpdateMessage(fmt.Sprintf("%s ✗ (schema not found)", fn))
							t.MarkAsErrored()
							return
						}

						var schema jtd.Schema
						if err := json.Unmarshal(schemaData, &schema); err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.UpdateMessage(fmt.Sprintf("%s ✗ (invalid schema)", fn))
							t.MarkAsErrored()
							return
						}

						var doc interface{}
						if err := json.Unmarshal([]byte(resp.RespBody), &doc); err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.UpdateMessage(fmt.Sprintf("%s ✗ (invalid response)", fn))
							t.MarkAsErrored()
							return
						}
						if validateErrors, err := jtd.Validate(schema, doc); len(validateErrors) != 0 || err != nil {
							fmt.Printf("%#v\n", validateErrors)
							mu.Lock()
							failed = true
							mu.Unlock()
							t.UpdateMessage(fmt.Sprintf("%s %d ✗", fn, resp.StatusCode))
							t.MarkAsErrored()
							return
						}

						// Success with status code
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
						t.Total = 100 // Switch to determinate progress
						t.UpdateMessage(statusColor.Sprintf("%s %d ✓", fn, resp.StatusCode))
						t.MarkAsDone()
					}(fileName, tracker)
				}
			}

			wg.Wait()

			// Wait for progress to finish rendering
			for pw.IsRenderInProgress() {
				if pw.LengthActive() == 0 {
					pw.Stop()
				}
			}

			if failed {
				fmt.Println("\nOne or more tests failed")
				os.Exit(1)
			}
			fmt.Println("\nAll tests passed")
		},
	}
}
