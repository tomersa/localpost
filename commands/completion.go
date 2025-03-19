package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Shell represents a supported shell type
type Shell string

// Supported shells as constants
const (
	ShellBash Shell = "bash"
	ShellZsh  Shell = "zsh"
	ShellFish Shell = "fish"
)

func NewCompletionCommand() *cobra.Command {
	var shell Shell
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Output completion script for your shell",
		Long: `Output the autocompletion script for the specified shell to stdout.
Supported shells: bash, zsh, fish.
Add 'source <(localpost completion --shell <shell>)' to your shell config file (e.g., ~/.zshrc, ~/.bashrc, ~/.config/fish/config.fish).
Example: 'source <(localpost completion --shell zsh)'`,
		Run: func(cmd *cobra.Command, args []string) {
			if shell == "" {
				fmt.Fprintf(os.Stderr, "Error: --shell flag is required (e.g., --shell zsh)\n")
				os.Exit(1)
			}

			var builder strings.Builder
			switch shell {
			case ShellBash:
				err := cmd.Root().GenBashCompletionV2(&builder, false)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error generating bash completion: %v\n", err)
					os.Exit(1)
				}
			case ShellZsh:
				// Add compinit initialization for Zsh
				builder.WriteString("# Ensure Zsh completion system is initialized\n")
				builder.WriteString("autoload -Uz compinit\n")
				builder.WriteString("compinit\n")
				err := cmd.Root().GenZshCompletion(&builder)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error generating zsh completion: %v\n", err)
					os.Exit(1)
				}
			case ShellFish:
				err := cmd.Root().GenFishCompletion(&builder, false)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error generating fish completion: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Fprintf(os.Stderr, "Error: unsupported shell: %s (use bash, zsh, or fish)\n", shell)
				os.Exit(1)
			}

			fmt.Print(builder.String())
		},
	}
	cmd.Flags().StringVar((*string)(&shell), "shell", "", "Specify shell for completion (bash, zsh, fish) [required]")
	cmd.MarkFlagRequired("shell") // Enforce --shell requirement
	return cmd
}
