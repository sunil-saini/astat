package route53

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Route53 hosted zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "route53-zones")
	},
}

func init() {
	Route53Cmd.AddCommand(listCmd)
}
