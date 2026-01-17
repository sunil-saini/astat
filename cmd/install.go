package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install astat",
	Long:    "Install astat including binary placement and autocomplete",
	GroupID: "project",
	Run: func(cmd *cobra.Command, args []string) {
		ui := NewUI()

		ui.Header("Welcome to astat!")
		ui.Println()

		// 1. Binary path check
		spinner := ui.Spinner("Checking environment...")
		_, err := os.Executable()
		if err != nil {
			spinner.Fail("Could not determine current executable path")
			return
		}
		spinner.Success("Environment check complete")

		// 2. Autocomplete installation
		shellPath := os.Getenv("SHELL")
		if shellPath != "" {
			shell := filepath.Base(shellPath)
			if shell == "zsh" || shell == "bash" || shell == "fish" {
				spinner = ui.Spinner(fmt.Sprintf("Setting up autocomplete for %s...", shell))
				installCompletionCmd.Run(cmd, []string{})
				spinner.Success(fmt.Sprintf("Autocomplete configured for %s", shell))
			} else {
				ui.Warning(fmt.Sprintf("Automatic autocomplete setup for %s is not supported yet", shell))
			}
		} else {
			ui.Warning("Could not detect shell for autocomplete setup")
		}

		// 3. Config initialization
		spinner = ui.Spinner("Initializing configuration...")
		if msg := installConfig(ui); msg != "" {
			spinner.Success(msg)
		} else {
			spinner.Success("Configuration initialized")
		}

		// 4. Initial refresh
		ui.Println()
		refreshCmd.Run(cmd, []string{})

		ui.Println()
		ui.Success("Installation complete! ✨")
		ui.Println()

		ui.Section("Next Steps")
		ui.BulletList([]string{
			"Run 'astat status' to check cache status and updates",
			"Run 'astat ec2 list' to see your instances",
			"Restart your terminal to enable autocomplete",
		})

		ui.Println()
	},
}

func installConfig(ui UI) string {
	home, err := os.UserHomeDir()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get home directory: %v", err))
		return ""
	}

	configDir := filepath.Join(home, ".config", "astat")
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			ui.Error(fmt.Sprintf("Failed to create config directory: %v", err))
			return ""
		}
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		settings := map[string]any{
			"ttl":                 viper.GetDuration("ttl").String(),
			"auto-refresh":        viper.GetBool("auto-refresh"),
			"output":              viper.GetString("output"),
			"route53-max-records": viper.GetInt("route53-max-records"),
		}

		defaultConfig, err := yaml.Marshal(settings)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to marshal default config: %v", err))
			return ""
		}

		if err := os.WriteFile(configFile, defaultConfig, 0644); err != nil {
			ui.Error(fmt.Sprintf("Failed to create default config file: %v", err))
			return ""
		}
		return fmt.Sprintf("Created default configuration at %s", configFile)
	}
	return ""
}
