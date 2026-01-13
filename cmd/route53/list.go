package route53

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/model"
	"github.com/sunil-saini/astat/internal/refresh"
	"github.com/sunil-saini/astat/internal/render"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Route53 hosted zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "route53-zones")

		var zones []model.Route53HostedZone
		hit, err := cache.Load(cacheFile, &zones)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "route53-zones", aws.FetchHostedZones)
			hit, err = cache.Load(cacheFile, &zones)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(zones))
		for _, z := range zones {
			rows = append(rows, []string{
				z.ID,
				z.Name,
			})
		}
		return render.Print(render.TableData{
			Headers: []string{"ID", "Name"},
			Rows:    rows,
			JSON:    zones,
		})
	},
}

func init() {
	Route53Cmd.AddCommand(listCmd)
}
