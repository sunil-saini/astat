package route53

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/render"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "List all Route53 records across all hosted zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, "route53-records", aws.FetchAllRoute53Records)
	},
}

func init() {
	Route53Cmd.AddCommand(recordsCmd)
}
