package lambda

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all Lambda functions with their details",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "lambda")
	},
}

func init() {
	LambdaCmd.AddCommand(listCmd)
}
