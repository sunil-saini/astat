package elb

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Load Balancers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, "elb", aws.FetchLoadBalancers)
	},
}

func init() {
	ElbCmd.AddCommand(listCmd)
}
