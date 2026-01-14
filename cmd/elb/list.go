package elb

import "github.com/spf13/cobra"

var ElbCmd = &cobra.Command{
	Use:     "elb",
	Short:   "Elastic Load Balancers",
	GroupID: "resources",
}
