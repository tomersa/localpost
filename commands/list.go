package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all requests",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) >= 1 {
				fmt.Println("Error: too many arguments")
				os.Exit(1)
			}
			var requestPaths []string
			filepath.Walk(util.RequestsDir, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
					relPath, err := filepath.Rel(util.RequestsDir, path)
					if err != nil {
						return err
					}
					relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
					requestPaths = append(requestPaths, strings.TrimSuffix(relPath, ".yaml"))
				}
				return nil
			})
			fmt.Println(strings.Join(requestPaths, "\n"))
		},
	}
}
