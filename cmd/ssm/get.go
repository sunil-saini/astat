package ssm

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/logger"
)

var getCmd = &cobra.Command{
	Use:   "get <parameter-name>",
	Short: "Get SSM parameter value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		cfg, err := aws.LoadConfig(ctx)
		if err != nil {
			logger.Error("AWS config load failed: %v", err)
			return err
		}

		val, err := aws.GetSSMParameter(ctx, cfg, args[0])
		if err != nil {
			logger.Error("Failed to get parameter %s: %v", args[0], err)
			return err
		}

		fmt.Println(val)
		return nil
	},
}

func init() {
	SSMCmd.AddCommand(getCmd)
}
