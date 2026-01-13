package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/logger"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage astat configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		configDir := filepath.Join(os.Getenv("HOME"), ".config", "astat")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				logger.Error("failed to create config directory: %v", err)
				return
			}
		}

		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			configFile = filepath.Join(configDir, "config.yaml")
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			logger.Error("failed to save config: %v", err)
			return
		}

		logger.Success("Set %s to %s", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		if len(settings) == 0 {
			logger.Info("No configuration set")
			return
		}

		fmt.Println("Current Configuration:")
		for k, v := range settings {
			fmt.Printf("  %s: %v\n", k, v)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(configSetCmd)
	ConfigCmd.AddCommand(configListCmd)
}
