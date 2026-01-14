package route53

import "github.com/spf13/cobra"

var Route53Cmd = &cobra.Command{
	Use:     "route53",
	Short:   "Route53 hosted zones and records",
	GroupID: "resources",
}
