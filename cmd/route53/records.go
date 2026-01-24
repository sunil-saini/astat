package route53

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "List all Route53 records across all hosted zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "route53-records")
	},
}

func init() {
	Route53Cmd.AddCommand(recordsCmd)
}
