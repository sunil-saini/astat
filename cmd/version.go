package cmd

import (
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultBigText.WithLetters(
			putils.LettersFromString("astat"),
		).Render()

		pterm.DefaultSection.Println("AWS Stats Indexer")

		data := pterm.TableData{
			{"Version", pterm.Cyan(version.Version)},
			{"Commit", pterm.Cyan(version.Commit)},
			{"Built Date", pterm.Cyan(version.Date)},
		}

		pterm.DefaultTable.WithData(data).Render()
	},
}
