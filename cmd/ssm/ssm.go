package ssm

import "github.com/spf13/cobra"

var SSMCmd = &cobra.Command{
	Use:     "ssm",
	Short:   "SSM parameter store",
	GroupID: "resources",
}
