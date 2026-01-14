package elb

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
	Short:   "List Load Balancers",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "elb")

		var lbs []model.LoadBalancer
		hit, err := cache.Load(cacheFile, &lbs)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "elb", aws.FetchLoadBalancers)
			hit, err = cache.Load(cacheFile, &lbs)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(lbs))
		for _, lb := range lbs {
			rows = append(rows, []string{
				lb.Type,
				lb.Name,
				lb.Scheme,
				lb.CreatedAt,
				lb.DNSName,
			})
		}
		return render.Print(render.TableData{
			Headers: []string{"Type", "Name", "Scheme", "Created At", "DNS"},
			Rows:    rows,
			JSON:    lbs,
		})
	},
}

func init() {
	ElbCmd.AddCommand(listCmd)
}
