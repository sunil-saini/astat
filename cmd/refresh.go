package cmd

import (
	"context"
	"sync"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
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
		ui := NewUI()

		ui.Println()
		ui.Info("Refreshing all services...")
		ui.Println()

		ui.StartRefresh()
		defer ui.StopRefresh()

		var wg sync.WaitGroup
		for _, svc := range registry.Registry {
			wg.Add(1)
			go func(service registry.Service) {
				defer wg.Done()
				tracker := ui.GetRefreshTracker(service.Name)
				refresh.Refresh(ctx, service.Name, func(ctx context.Context, cfg sdkaws.Config) (any, error) {
					return service.Fetch(ctx, cfg)
				}, tracker)
			}(svc)
		}
		wg.Wait()

		ui.Println()
		ui.Success("All services refreshed!")
	},
}
