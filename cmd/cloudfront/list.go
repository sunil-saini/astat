package cloudfront

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
	Short:   "List CloudFront distributions",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "cloudfront")

		var dists []model.CloudFrontDistribution
		hit, err := cache.Load(cacheFile, &dists)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "cloudfront", aws.FetchCloudFront)
			hit, err = cache.Load(cacheFile, &dists)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(dists))
		for _, d := range dists {
			rows = append(rows, []string{
				d.ID,
				d.Domain,
				d.Status,
			})
		}
		return render.Print(render.TableData{
			Headers: []string{"ID", "Domain", "Status"},
			Rows:    rows,
			JSON:    dists,
		})
	},
}

func init() {
	CloudFrontCmd.AddCommand(listCmd)
}
