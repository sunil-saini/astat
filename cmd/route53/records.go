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

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "List all Route53 records across all hosted zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "route53-records")

		var records []model.Route53Record
		hit, err := cache.Load(cacheFile, &records)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "route53-records", aws.FetchAllRoute53Records)
			hit, err = cache.Load(cacheFile, &records)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(records))
		for _, r := range records {
			rows = append(rows, []string{
				r.ZoneName,
				r.Name,
				r.Type,
				r.TTL,
				r.Value,
			})
		}

		return render.Print(render.TableData{
			Headers: []string{"Zone", "Name", "Type", "TTL", "Value"},
			Rows:    rows,
			JSON:    records,
		})
	},
}

func init() {
	Route53Cmd.AddCommand(recordsCmd)
}
