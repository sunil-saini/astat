package ssm

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List SSM parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, "ssm", aws.FetchSSMParameters)
	},
}

func init() {
	SSMCmd.AddCommand(listCmd)
}
