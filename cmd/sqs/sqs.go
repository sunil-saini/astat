package sqs

import "github.com/spf13/cobra"

var SQSCmd = &cobra.Command{
	Use:     "sqs",
	Short:   "SQS Queues",
	GroupID: "resources",
}
