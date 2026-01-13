package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/logger"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(astat completion bash)

  # To load completions for each session, add to the end of your ~/.bashrc:

  $ astat completion bash > /etc/bash_completion.d/astat

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, add to the end of your ~/.zshrc:

  $ astat completion zsh > "${fpath[1]}/_astat"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ astat completion fish > ~/.config/fish/completions/astat.fish

PowerShell:

  PS> astat completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:

  PS> astat completion powershell > astat.ps1
  PS> Add-Content $PROFILE.CurrentUserCurrentHost -Value (Get-Content astat.ps1)
  PS> Remove-Item astat.ps1
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

var installCompletionCmd = &cobra.Command{
	Use:   "install",
	Short: "Automatically install autocomplete script for your current shell",
	Run: func(cmd *cobra.Command, args []string) {
		shellPath := os.Getenv("SHELL")
		if shellPath == "" {
			logger.Error("Could not detect SHELL environment variable")
			return
		}

		shell := filepath.Base(shellPath)
		home, _ := os.UserHomeDir()

		switch shell {
		case "zsh":
			rcPath := filepath.Join(home, ".zshrc")
			installToShell(rcPath, `(( $+commands[astat] )) && source <(astat completion zsh) 2>/dev/null`)
		case "bash":
			rcPath := filepath.Join(home, ".bashrc")
			installToShell(rcPath, `command -v astat >/dev/null 2>&1 && source <(astat completion bash)`)
		default:
			logger.Warn("Automatic installation for %s is not supported yet. Please use manual instructions from 'astat completion --help'", shell)
		}
	},
}

func installToShell(rcPath, line string) {
	content, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		logger.Error("Failed to read %s: %v", rcPath, err)
		return
	}

	if strings.Contains(string(content), line) {
		logger.Info("Autocomplete already installed in %s", rcPath)
		return
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("Failed to open %s for writing: %v", rcPath, err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString("\n# astat autocomplete\n" + line + "\n"); err != nil {
		logger.Error("Failed to write to %s: %v", rcPath, err)
		return
	}

	logger.Success("Autocomplete installed in %s. Please restart your shell or run: source %s", rcPath, rcPath)
}

func init() {
	completionCmd.AddCommand(installCompletionCmd)
}
