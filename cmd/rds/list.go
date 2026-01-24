package rds

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all RDS clusters with their details",
	Long: `List all RDS clusters with their details

Examples:
  # List all clusters
  astat rds list

  # Force refresh from AWS
  astat rds list --refresh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "rds-clusters")
	},
}

func init() {
	RDSCmd.AddCommand(listCmd)
}
