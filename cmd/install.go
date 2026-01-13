package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/logger"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install astat",
	Long:  "Install astat including binary placement and autocomplete.",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.WithFullWidth().Println("Welcome to astat!")
		fmt.Println()

		// 1. Binary path check
		spinner, _ := pterm.DefaultSpinner.Start("Checking environment...")
		_, err := os.Executable()
		if err != nil {
			spinner.Fail("Could not determine current executable path")
			return
		}

		spinner.UpdateText("Installing binary to /usr/local/bin...")
		installBinary()

		// 2. Autocomplete installation
		spinner.UpdateText("Setting up autocomplete...")
		shellPath := os.Getenv("SHELL")
		if shellPath != "" {
			shell := filepath.Base(shellPath)
			if shell == "zsh" || shell == "bash" {
				installCompletionCmd.Run(cmd, []string{})
				spinner.Success(fmt.Sprintf("Autocomplete configured for %s", shell))
			} else {
				spinner.Warning(fmt.Sprintf("Automatic autocomplete setup for %s is not supported yet", shell))
			}
		} else {
			spinner.Warning("Could not detect shell for autocomplete setup")
		}

		fmt.Println()
		pterm.Success.Println("Installation complete! ✨")
		fmt.Println()

		pterm.DefaultSection.Println("Next Steps")
		pterm.BulletListPrinter{
			Items: []pterm.BulletListItem{
				{Text: "Run " + pterm.Cyan("astat ec2 list") + " to see your instances", Bullet: "→"},
				{Text: "Run " + pterm.Cyan("astat config list") + " to see current settings", Bullet: "→"},
				{Text: "Restart your terminal to enable autocomplete", Bullet: "→"},
			},
		}.Render()

		fmt.Println()
	},
}

func installBinary() {
	if runtime.GOOS == "windows" {
		logger.Warn("Windows is not supported yet!")
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		logger.Error("Could not find current binary: %v", err)
		return
	}

	target := "/usr/local/bin/astat"

	// Check if already installed
	if execPath == target {
		return
	}

	logger.Info("Attempting to install astat to %s...", target)

	// Check if we have write access to /usr/local/bin
	if err := os.WriteFile(target+"_test", []byte("test"), 0644); err != nil {
		logger.Warn("Permission denied for /usr/local/bin. You might need to run: sudo cp %s %s", execPath, target)
		return
	}
	os.Remove(target + "_test")

	input, err := os.ReadFile(execPath)
	if err != nil {
		logger.Error("Failed to read current binary: %v", err)
		return
	}

	err = os.WriteFile(target, input, 0755)
	if err != nil {
		logger.Error("Failed to copy binary: %v", err)
		return
	}

	logger.Success("Installed astat to %s", target)
}
