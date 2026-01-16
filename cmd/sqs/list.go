package sqs

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
	Short:   "List all SQS queues",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "sqs")

		var queues []model.SQSQueue
		hit, err := cache.Load(cacheFile, &queues)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "sqs", aws.FetchSQSQueues)
			hit, err = cache.Load(cacheFile, &queues)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(queues))
		for _, q := range queues {
			rows = append(rows, []string{
				q.Name,
				q.Type,
			})
		}

		return render.Print(render.TableData{
			Headers: []string{"Name", "Type"},
			Rows:    rows,
			JSON:    queues,
		})
	},
}

func init() {
	SQSCmd.AddCommand(listCmd)
}
