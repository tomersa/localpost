package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate and install completion script for your shell",
		Long: `Generate the autocompletion script for the specified shell and install it to ~/.localpost-completions with automatic sourcing.
Supported shells: bash, zsh, fish`,
		Run: func(cmd *cobra.Command, args []string) {
			shell := detectShell()
			if shell == "" {
				fmt.Println("Error: unsupported or undetectable shell")
				os.Exit(1)
			}

			completionDir := filepath.Join(os.Getenv("HOME"), ".localpost-completions")
			var completionFile, configFile, sourceLine string
			var builder strings.Builder

			switch shell {
			case "bash":
				err := cmd.Root().GenBashCompletionV2(&builder, false)
				completionFile = filepath.Join(completionDir, "_bash")
				configFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
				sourceLine = fmt.Sprintf("source %s", completionFile)
				if err != nil {
					fmt.Printf("Error generating bash completion: %v\n", err)
					os.Exit(1)
				}
			case "zsh":
				err := cmd.Root().GenZshCompletion(&builder) // Use GenZshCompletion for richer completion
				completionFile = filepath.Join(completionDir, "_zsh")
				configFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
				// Add fpath and menu styling
				sourceLine = fmt.Sprintf(`[[ -d %s ]] && fpath+=(%s)
autoload -Uz compinit && compinit -u
zstyle ':completion:*' menu select
zstyle ':completion:*' list-colors ''
`, completionDir, completionDir)
				if err != nil {
					fmt.Printf("Error generating zsh completion: %v\n", err)
					os.Exit(1)
				}
			case "fish":
				err := cmd.Root().GenFishCompletion(&builder, false)
				completionFile = filepath.Join(completionDir, "_fish")
				configFile = filepath.Join(os.Getenv("HOME"), ".config", "fish", "config.fish")
				sourceLine = fmt.Sprintf("source %s", completionFile)
				if err != nil {
					fmt.Printf("Error generating fish completion: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("Error: unsupported shell: %s\n", shell)
				os.Exit(1)
			}

			completionScript := builder.String()

			// Debug: Print the generated script
			fmt.Printf("Generated %s completion script:\n%s\n", shell, completionScript)

			if err := os.MkdirAll(completionDir, 0755); err != nil {
				fmt.Printf("Error creating completion directory: %v\n", err)
				os.Exit(1)
			}

			if err := os.WriteFile(completionFile, []byte(completionScript), 0644); err != nil {
				fmt.Printf("Error writing completion file: %v\n", err)
				os.Exit(1)
			}

			configDir := filepath.Dir(configFile)
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Printf("Error creating config directory: %v\n", err)
				os.Exit(1)
			}

			if err := appendToConfig(configFile, sourceLine); err != nil {
				fmt.Printf("Error updating %s: %v\n", configFile, err)
				os.Exit(1)
			}

			fmt.Printf("Installed %s completion script to %s\n", shell, completionFile)
			fmt.Printf("Updated %s to source the completion script automatically.\n", configFile)
			fmt.Printf("Restart your shell or run 'source %s' to apply changes immediately.\n", configFile)
		},
	}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	return ""
}

func appendToConfig(configFile, sourceLine string) error {
	content, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading config file: %v", err)
	}

	// Check for key parts to avoid duplicates
	if strings.Contains(string(content), sourceLine) || (strings.Contains(string(content), fmt.Sprintf("fpath+=(%s)", filepath.Dir(configFile))) && strings.Contains(string(content), "zstyle ':completion:*' menu select")) {
		return nil
	}

	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + sourceLine + "\n"); err != nil {
		return fmt.Errorf("error writing to config file: %v", err)
	}

	return nil
}
