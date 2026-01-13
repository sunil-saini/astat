package lambda

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
	Short:   "List all Lambda functions with their details",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "lambda")

		var funcs []model.LambdaFunction
		hit, err := cache.Load(cacheFile, &funcs)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "lambda", aws.FetchLambdaFunctions)
			hit, err = cache.Load(cacheFile, &funcs)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(funcs))
		for _, f := range funcs {
			rows = append(rows, []string{
				f.Name,
				f.Runtime,
				f.LastModified,
				f.Memory,
				f.Timeout,
			})
		}
		return render.Print(render.TableData{
			Headers: []string{"Name", "Runtime", "Last Modified", "Memory (MB)", "Timeout (s)"},
			Rows:    rows,
			JSON:    funcs,
		})
	},
}

func init() {
	LambdaCmd.AddCommand(listCmd)
}
