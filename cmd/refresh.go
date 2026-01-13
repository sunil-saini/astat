package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/refresh"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh all services",
	Long:  "Refresh cache for all AWS services",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		services := []struct {
			name string
			fn   func(context.Context, *pterm.MultiPrinter)
		}{
			{"ec2", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "ec2", aws.FetchEC2Instances, multi)
			}},
			{"s3", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "s3", aws.FetchS3Buckets, multi)
			}},
			{"lambda", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "lambda", aws.FetchLambdaFunctions, multi)
			}},
			{"cloudfront", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "cloudfront", aws.FetchCloudFront, multi)
			}},
			{"route53-zones", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "route53-zones", aws.FetchHostedZones, multi)
			}},
			{"route53-records", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "route53-records", aws.FetchAllRoute53Records, multi)
			}},
			{"ssm", func(ctx context.Context, multi *pterm.MultiPrinter) {
				refresh.RefreshWithMulti(ctx, "ssm", aws.FetchSSMParameters, multi)
			}},
		}

		multi := pterm.DefaultMultiPrinter
		multi.Start()
		defer multi.Stop()

		fmt.Println()
		pterm.Info.Println("Refreshing all services...")
		fmt.Println()

		var wg sync.WaitGroup
		for _, svc := range services {
			wg.Add(1)
			go func(s struct {
				name string
				fn   func(context.Context, *pterm.MultiPrinter)
			}) {
				defer wg.Done()
				s.fn(ctx, &multi)
			}(svc)
		}
		wg.Wait()

		fmt.Println()
		pterm.Success.Println("All services refreshed!")
	},
}
