package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("astat version %s\n", version.Version)
		fmt.Printf("commit: %s\n", version.Commit)
		fmt.Printf("built date: %s\n", version.Date)
	},
}
