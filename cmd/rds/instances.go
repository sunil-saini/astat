package rds

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

var instancesCmd = &cobra.Command{
	Use:     "instances",
	Aliases: []string{"ins"},
	Short:   "List all RDS instances with their details",
	Long: `List all RDS instances with their details

Examples:
  # List all instances
  astat rds instances

  # Force refresh from AWS
  astat rds instances --refresh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "rds-instances")

		var instances []model.RDSInstance
		hit, err := cache.Load(cacheFile, &instances)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "rds-instances", aws.FetchRDSInstances)
			hit, err = cache.Load(cacheFile, &instances)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(instances))
		for _, i := range instances {
			rows = append(rows, []string{
				i.ClusterIdentifier,
				i.InstanceIdentifier,
				i.Role,
				i.Engine,
				i.EngineVersion,
				i.DBInstanceStatus,
				i.InstanceClass,
				i.AvailabilityZone,
			})
		}

		return render.Print(render.TableData{
			Headers: []string{
				"Cluster", "Identifier", "Role", "Engine", "Engine Version", "Status", "Class", "AZ",
			},
			Rows: rows,
			JSON: instances,
		})
	},
}

func init() {
	RDSCmd.AddCommand(instancesCmd)
}
