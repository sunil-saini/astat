package ec2

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all EC2 instances with their details",
	Long: `List all EC2 instances with their details

Examples:
  # List all instances
  astat ec2 list

  # Force refresh from AWS
  astat ec2 list --refresh

  # Output as JSON
  astat ec2 list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "ec2")
	},
}

func init() {
	EC2Cmd.AddCommand(listCmd)
}
