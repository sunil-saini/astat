package cloudfront

import "github.com/spf13/cobra"

var CloudFrontCmd = &cobra.Command{
	Use:     "cloudfront",
	Short:   "CloudFront distributions",
	GroupID: "resources",
}
