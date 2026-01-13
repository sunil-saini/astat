package ssm

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
	Short:   "List SSM parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "ssm")

		var params []model.SSMParameter
		hit, err := cache.Load(cacheFile, &params)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "ssm", aws.FetchSSMParameters)
			hit, err = cache.Load(cacheFile, &params)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(params))
		for _, p := range params {
			rows = append(rows, []string{
				p.Name,
				p.Type,
			})
		}
		return render.Print(render.TableData{
			Headers: []string{"Name", "Type"},
			Rows:    rows,
			JSON:    params,
		})
	},
}

func init() {
	SSMCmd.AddCommand(listCmd)
}
