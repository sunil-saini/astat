package rds

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var instancesCmd = &cobra.Command{
	Use:     "instances",
	Aliases: []string{"ins"},
	Short:   "List all RDS instances with their details",
	Long: `List all RDS instances with their details

Examples:
  # List all instances
  astat rds instances

  # Force refresh from AWS
  astat rds instances --refresh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "rds-instances")
	},
}

func init() {
	RDSCmd.AddCommand(instancesCmd)
}
