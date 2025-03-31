package commands

import (
	"encoding/json"
	"fmt"
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
			// Clear cookies
			if err := util.ClearCookies(); err != nil {
				fmt.Printf("Error clearing cookies: %v\n", err)
				os.Exit(1)
			}

			// Execute login request
			env, err := util.LoadEnv()
			if err != nil {
				fmt.Printf("Error loading env: %v\n", err)
				os.Exit(1)
			}
			if env.Login != nil && env.Login.Request != "" {
				_, err := util.HandleRequest(env.Login.Request)
				if err != nil {
					fmt.Printf("Error executing login request %s: %v\n", env.Login.Request, err)
					os.Exit(1)
				}
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

			go pw.Render()

			// Track failures
			var wg sync.WaitGroup
			failed := false
			mu := sync.Mutex{}

			files, err := os.ReadDir(util.RequestsDir)
			if err != nil {
				fmt.Printf("Error reading requests dir: %v\n", err)
				os.Exit(1)
			}

			for _, file := range files {
				fileName := removeExtension(file.Name())
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
					if env.Login != nil && fileName == strings.TrimSuffix(env.Login.Request, ".yaml") {
						continue
					}

					wg.Add(1)
					tracker := &progress.Tracker{
						Message: fileName,
						Total:   100,
					}
					pw.AppendTracker(tracker)

					go func(t *progress.Tracker, fn string) {
						defer wg.Done()

						resp, err := util.HandleRequest(fn)
						if err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.MarkAsErrored()
							return
						}

						schemaPath := filepath.Join(util.SchemasDir, fn+".jtd.json")
						schemaData, err := os.ReadFile(schemaPath)
						if err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.MarkAsErrored()
							return
						}

						var schema jtd.Schema
						if err := json.Unmarshal(schemaData, &schema); err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.MarkAsErrored()
							return
						}

						var doc interface{}
						if err := json.Unmarshal([]byte(resp.RespBody), &doc); err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.MarkAsErrored()
							return
						}
						if _, _ = jtd.Validate(schema, doc); err != nil {
							mu.Lock()
							failed = true
							mu.Unlock()
							t.MarkAsErrored()
							return
						}

						t.MarkAsDone()
					}(tracker, fileName)
				}
			}

			wg.Wait()

			for pw.IsRenderInProgress() {
				if pw.LengthActive() == 0 {
					pw.Stop()
				}
				time.Sleep(time.Millisecond * 100)
			}

			if failed {
				fmt.Println("\nOne or more tests failed")
				os.Exit(1)
			}
			fmt.Println("\nAll tests passed")
		},
	}
}
