package domain

import (
	"github.com/spf13/cobra"
)

var DomainCmd = &cobra.Command{
	Use:     "domain",
	Short:   "Domain and DNS related tools",
	GroupID: "resources",
}
