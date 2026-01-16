package cmd

import (
	"context"
	"fmt"
	"sync"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/model"
	"github.com/sunil-saini/astat/internal/refresh"
	"github.com/sunil-saini/astat/internal/registry"
)

var refreshCmd = &cobra.Command{
	Use:     "refresh",
	Short:   "Refresh all services",
	Long:    "Refresh cache for all AWS services",
	GroupID: "project",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		multi := pterm.DefaultMultiPrinter
		multi.Start()
		defer multi.Stop()

		fmt.Println()
		pterm.Info.Println("Refreshing all services...")
		fmt.Println()

		var wg sync.WaitGroup
		for _, svc := range registry.Registry {
			wg.Add(1)
			go func(s registry.Service) {
				defer wg.Done()
				refresh.RefreshWithMulti(ctx, s.Name, func(ctx context.Context, cfg sdkaws.Config) ([]any, error) {
					res, err := s.Fetch(ctx, cfg)
					if err != nil {
						return nil, err
					}
					return model.ToAnySlice(res), nil
				}, &multi)
			}(svc)
		}
		wg.Wait()

		fmt.Println()
		pterm.Success.Println("All services refreshed!")
	},
}
