package rds

import "github.com/spf13/cobra"

var RDSCmd = &cobra.Command{
	Use:     "rds",
	Short:   "RDS Clusters and Instances",
	GroupID: "resources",
}
