package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/refresh"
	"github.com/sunil-saini/astat/internal/version"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache status and check for updates",
	Long: `Display the status of all cached AWS services.

Shows when each service was last refreshed and whether
the cache is fresh or stale based on the configured TTL.
Also checks for available astat updates.

Examples:
  # Check cache status
  astat status`,
	Run: func(cmd *cobra.Command, args []string) {
		var meta cache.Meta
		metaPath := cache.Path(cache.Dir(), "meta")
		cacheInitialized := false
		if err := cache.Read(metaPath, &meta); err != nil {
			logger.Warn("Cache metadata not found, initializing with defaults...")
			meta = cache.Meta{
				LastUpdated: time.Time{},
				Services:    make(map[string]cache.ServiceMeta),
			}

			if err := cache.EnsureDir(cache.Dir()); err != nil {
				logger.Error("Failed to create cache directory: %v", err)
				return
			}

			if err := cache.Write(metaPath, meta); err != nil {
				logger.Error("Failed to initialize cache metadata: %v", err)
				return
			}
			logger.Success("Cache initialized successfully")
			cacheInitialized = true

			logger.Info("Triggering initial refresh to populate cache...")
			refreshCmd.Run(cmd, args)

			if err := cache.Read(metaPath, &meta); err != nil {
				logger.Error("Failed to read cache metadata after refresh: %v", err)
				return
			}
		}

		ttl := viper.GetDuration("ttl")
		allServices := []string{"ec2", "s3", "lambda", "cloudfront", "route53-zones", "route53-records", "ssm"}

		logger.Info("Overall Last Refresh: %v", meta.LastUpdated.Format(time.RFC1123))
		logger.Info("TTL: %v", ttl)

		isAnyStale := false
		for _, s := range allServices {
			sMeta, ok := meta.Services[s]
			if !ok {
				logger.Warn("%-12s: NEVER REFRESHED", s)
				isAnyStale = true
				continue
			}

			if sMeta.Refreshing && refresh.IsProcessAlive(sMeta.BusyPID) {
				logger.Info("%-12s: REFRESHING (PID: %d)", s, sMeta.BusyPID)
				continue
			}

			age := time.Since(sMeta.LastUpdated)
			if age > ttl {
				logger.Warn("%-12s: STALE (%v ago)", s, age.Truncate(time.Second))
				isAnyStale = true
			} else {
				logger.Success("%-12s: FRESH (%v ago)", s, age.Truncate(time.Second))
			}
		}

		if isAnyStale && !cacheInitialized {
			if viper.GetBool("auto-refresh") {
				logger.Info("Auto-refresh is enabled. Stale services will be updated on next command.")
			}
		}

		fmt.Println()
		available, latestVersion, _, err := version.IsUpgradeAvailable()
		if err == nil && available {
			logger.Warn("New version available: %s (current: %s)", latestVersion, version.Version)
			logger.Info("Run 'astat upgrade' to update")
		}
	},
}
