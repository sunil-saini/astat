package lambda

import "github.com/spf13/cobra"

var LambdaCmd = &cobra.Command{
	Use:     "lambda",
	Short:   "Lambda functions",
	GroupID: "resources",
}
