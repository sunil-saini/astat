package rds

import (
	"fmt"

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
	Short:   "List all RDS clusters with their details",
	Long: `List all RDS clusters with their details

Examples:
  # List all clusters
  astat rds list

  # Force refresh from AWS
  astat rds list --refresh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "rds-clusters")

		var clusters []model.RDSCluster
		hit, err := cache.Load(cacheFile, &clusters)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "rds-clusters", aws.FetchRDSClusters)
			hit, err = cache.Load(cacheFile, &clusters)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(clusters))
		for _, c := range clusters {
			rows = append(rows, []string{
				c.ClusterIdentifier,
				c.Status,
				c.Engine,
				c.EngineVersion,
				c.MultiAZ,
				c.IsPublic,
				fmt.Sprintf("%d", c.InstanceCount),
				c.StorageType,
				c.CreateTime,
			})
		}

		return render.Print(render.TableData{
			Headers: []string{
				"Identifier", "Status", "Engine", "Engine Version", "Multi-AZ", "Public Access", "Instance Count", "Storage Type", "Created At",
			},
			Rows: rows,
			JSON: clusters,
		})
	},
}

func init() {
	RDSCmd.AddCommand(listCmd)
}
