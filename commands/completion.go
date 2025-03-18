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
	ShellNone Shell = ""
)

func NewCompletionCommand() *cobra.Command {
	var shellOverride Shell
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Output completion script for your shell",
		Long: `Output the autocompletion script for the specified shell to stdout.
Supported shells: bash, zsh, fish.
Add 'source <(localpost completion --shell <shell>)' to your shell config file (e.g., ~/.zshrc, ~/.bashrc, ~/.config/fish/config.fish).
Example: 'source <(localpost completion --shell zsh)'`,
		Run: func(cmd *cobra.Command, args []string) {
			shell := shellOverride
			if shell == ShellNone {
				shell = detectShell()
				if shell == ShellNone {
					fmt.Fprintf(os.Stderr, "Error: unsupported or undetectable shell; specify with --shell (bash, zsh, fish)\n")
					os.Exit(1)
				}
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
				fmt.Fprintf(os.Stderr, "Error: unsupported shell: %s\n", shell)
				os.Exit(1)
			}

			fmt.Print(builder.String())
		},
	}
	cmd.Flags().StringVar((*string)(&shellOverride), "shell", "", "Specify shell for completion (bash, zsh, fish)")
	return cmd
}

func detectShell() Shell {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ShellNone
	}
	switch {
	case strings.Contains(shell, "bash"):
		return ShellBash
	case strings.Contains(shell, "zsh"):
		return ShellZsh
	case strings.Contains(shell, "fish"):
		return ShellFish
	default:
		return ShellNone
	}
}
