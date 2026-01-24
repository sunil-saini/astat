package s3

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all S3 buckets with their details",
	Long: `List all S3 buckets in your AWS account

Displays bucket name, creation date, and region.
Data is served from local cache for instant results.

Examples:
  # List all buckets
  astat s3 list
  
  # Force refresh from AWS
  astat s3 list --refresh
  
  # Output as JSON
  astat s3 list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "s3")
	},
}

func init() {
	S3Cmd.AddCommand(listCmd)
}
