package sqs

import (
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all SQS queues",
	RunE: func(cmd *cobra.Command, args []string) error {
		return render.List(cmd, args, "sqs")
	},
}

func init() {
	SQSCmd.AddCommand(listCmd)
}
