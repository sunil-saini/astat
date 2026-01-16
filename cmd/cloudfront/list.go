package cloudfront

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List CloudFront distributions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, "cloudfront", aws.FetchCloudFront)
	},
}

func init() {
	CloudFrontCmd.AddCommand(listCmd)
}
