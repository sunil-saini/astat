package s3

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
	Short:   "List all S3 buckets with their details",
	Long: `List all S3 buckets in your AWS account

Displays bucket name, creation date, and region.
Data is served from local cache for instant results.

Examples:
  # List all buckets
  astat s3 list
  
  # Force refresh from AWS
  astat s3 list --refresh
  
  # Output as JSON
  astat s3 list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "s3")

		var buckets []model.S3Bucket
		hit, err := cache.Load(cacheFile, &buckets)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "s3", aws.FetchS3Buckets)
			hit, err = cache.Load(cacheFile, &buckets)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(buckets))
		for _, b := range buckets {
			rows = append(rows, []string{
				b.Name,
				b.Region,
				b.CreationDate,
			})
		}
		return render.Print(render.TableData{
			Headers: []string{"Name", "Region", "Creation Date"},
			Rows:    rows,
			JSON:    buckets,
		})
	},
}

func init() {
	S3Cmd.AddCommand(listCmd)
}
