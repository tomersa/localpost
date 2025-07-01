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
	ShellBash       Shell = "bash"
	ShellZsh        Shell = "zsh"
	ShellFish       Shell = "fish"
	ShellPowerShell Shell = "powershell"
)

func CompletionCmd() *cobra.Command {
	var shell Shell
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Output completion script for your shell",
		Long: `Output the autocompletion script for the specified shell to stdout.
Supported shells: bash, zsh, fish, powershell.

For bash/zsh/fish, add 'source <(localpost completion --shell <shell>)' to your shell config file:
- Bash: ~/.bashrc
- Zsh: ~/.zshrc
- Fish: ~/.config/fish/config.fish
Example: 'source <(localpost completion --shell zsh)'

For powershell, run './setup-completion.ps1' to automate setup, or manually:
1. Save the output: 'localpost completion --shell powershell > localpost_completion.ps1'
2. Move to PowerShell profile directory: 'Move-Item localpost_completion.ps1 (Split-Path $PROFILE -Parent)'
3. Add to $PROFILE: '. (Join-Path (Split-Path $PROFILE -Parent) "localpost_completion.ps1")'
4. Reload profile: '. $PROFILE'`,
		Run: func(cmd *cobra.Command, args []string) {
			if shell == "" {
				fmt.Fprintf(os.Stderr, "Error: --shell flag is required (e.g., --shell powershell)\n")
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
				builder.WriteString("# Ensure Zsh completion system is initialized\n")
				builder.WriteString("autoload -Uz compinit && compinit\n")
				builder.WriteString("zstyle ':completion:*' menu select\nzstyle ':completion:*' list-colors ''\n")
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
			case ShellPowerShell:
				err := cmd.Root().GenPowerShellCompletion(&builder)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error generating powershell completion: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Fprintf(os.Stderr, "Error: unsupported shell: %s (use bash, zsh, fish, or powershell)\n", shell)
				os.Exit(1)
			}

			fmt.Print(builder.String())
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	cmd.Flags().StringVar((*string)(&shell), "shell", "", "Specify shell for completion (bash, zsh, fish, powershell) [required]")
	cmd.MarkFlagRequired("shell")
	return cmd
}
