package cmd

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
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
		allServices := []string{"ec2", "s3", "lambda", "cloudfront", "route53-zones", "route53-records", "ssm", "elb"}

		pterm.DefaultSection.Println("Cache Status")
		pterm.Printf("%s: %s\n", pterm.LightMagenta("Last Refresh"), pterm.Cyan(meta.LastUpdated.Format(time.RFC1123)))
		pterm.Printf("%s:          %s\n\n", pterm.LightMagenta("TTL"), pterm.Cyan(ttl))

		data := pterm.TableData{
			{"Service", "Status", "Age"},
			{"───────────────", "───────────────", "───────────────"},
		}

		isAnyStale := false
		for _, s := range allServices {
			sMeta, ok := meta.Services[s]
			if !ok {
				data = append(data, []string{s, pterm.LightRed("✗ NEVER"), pterm.LightRed("-")})
				isAnyStale = true
				continue
			}

			if sMeta.Refreshing && refresh.IsProcessAlive(sMeta.BusyPID) {
				data = append(data, []string{s, pterm.LightBlue("● REFRESHING"), pterm.LightBlue(fmt.Sprintf("PID: %d", sMeta.BusyPID))})
				continue
			}

			statusText := pterm.LightGreen("✓ FRESH")
			ageText := pterm.LightGreen("-")

			if !sMeta.LastUpdated.IsZero() {
				age := time.Since(sMeta.LastUpdated).Truncate(time.Second)
				ageText = pterm.LightGreen(age.String() + " ago")

				if age > ttl {
					statusText = pterm.LightYellow("⚠ STALE")
					ageText = pterm.LightYellow(age.String() + " ago")
					isAnyStale = true
				}
			} else {
				statusText = pterm.LightYellow("⚠ STALE")
				isAnyStale = true
			}

			data = append(data, []string{s, statusText, ageText})
		}

		pterm.DefaultTable.
			WithBoxed().
			WithHasHeader().
			WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
			WithData(data).
			Render()

		if isAnyStale && !cacheInitialized {
			if viper.GetBool("auto-refresh") {
				pterm.Println()
				pterm.Info.Println("Auto-refresh is enabled. Stale services will be updated on next command")
			}
		}

		available, latestVersion, _, err := version.IsUpgradeAvailable()
		if err == nil && available {
			pterm.Println()
			pterm.Warning.Printf("New version available: %s (current: %s)\n", pterm.Green(latestVersion), version.Version)
			pterm.Info.Println("Run 'astat upgrade' to update")
		}
	},
}
