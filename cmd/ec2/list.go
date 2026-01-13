package ec2

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
	Short:   "List all EC2 instances with their details",
	Long: `List all EC2 instances with their details

Examples:
  # List all instances
  astat ec2 list

  # Force refresh from AWS
  astat ec2 list --refresh

  # Output as JSON
  astat ec2 list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		forceRefresh := viper.GetBool("refresh")
		cacheFile := cache.Path(cache.Dir(), "ec2")

		var instances []model.EC2Instance
		hit, err := cache.Load(cacheFile, &instances)
		if err != nil {
			logger.Error("cache read failed: %v", err)
			return err
		}

		ctx := cmd.Context()

		if !hit || forceRefresh {
			refresh.RefreshSync(ctx, "ec2", aws.FetchEC2Instances)
			hit, err = cache.Load(cacheFile, &instances)
			if err != nil {
				return err
			}
		}

		rows := make([][]string, 0, len(instances))
		for _, i := range instances {
			rows = append(rows, []string{
				i.InstanceID,
				i.Name,
				i.State,
				i.InstanceType,
				i.AZ,
				i.PrivateIP,
				i.PublicIP,
			})
		}

		return render.Print(render.TableData{
			Headers: []string{
				"ID", "Name", "State", "Type", "AZ", "Private IP", "Public IP",
			},
			Rows: rows,
			JSON: instances,
		})
	},
}

func init() {
	EC2Cmd.AddCommand(listCmd)
}
