package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/logger"
	"gopkg.in/yaml.v3"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install astat",
	Long:    "Install astat including binary placement and autocomplete",
	GroupID: "project",
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

		// 3. Config initialization
		spinner.UpdateText("Initializing configuration...")
		installConfig()
		spinner.Success("Configuration initialized")

		// 4. Initial refresh
		fmt.Println()
		refreshCmd.Run(cmd, []string{})

		fmt.Println()
		pterm.Success.Println("Installation complete! ✨")
		fmt.Println()

		pterm.DefaultSection.Println("Next Steps")
		pterm.BulletListPrinter{
			Items: []pterm.BulletListItem{
				{Text: "Run " + pterm.Cyan("astat status") + " to check cache status and updates", Bullet: "→"},
				{Text: "Run " + pterm.Cyan("astat ec2 list") + " to see your instances", Bullet: "→"},
				{Text: "Restart your terminal to enable autocomplete", Bullet: "→"},
			},
		}.Render()

		fmt.Println()
	},
}

func installConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory: %v", err)
		return
	}

	configDir := filepath.Join(home, ".config", "astat")
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			logger.Error("Failed to create config directory: %v", err)
			return
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
			logger.Error("Failed to marshal default config: %v", err)
			return
		}

		if err := os.WriteFile(configFile, defaultConfig, 0644); err != nil {
			logger.Error("Failed to create default config file: %v", err)
			return
		}
		logger.Success("Created default configuration at %s", configFile)
	}
}
