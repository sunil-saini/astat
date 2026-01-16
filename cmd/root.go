package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/cmd/cloudfront"
	"github.com/sunil-saini/astat/cmd/domain"
	"github.com/sunil-saini/astat/cmd/ec2"
	"github.com/sunil-saini/astat/cmd/elb"
	"github.com/sunil-saini/astat/cmd/lambda"
	"github.com/sunil-saini/astat/cmd/rds"
	"github.com/sunil-saini/astat/cmd/route53"
	"github.com/sunil-saini/astat/cmd/s3"
	"github.com/sunil-saini/astat/cmd/sqs"
	"github.com/sunil-saini/astat/cmd/ssm"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/refresh"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "astat",
	Short: "⚡ Lightning fast local AWS stats indexer",
	Long: `astat - AWS Stats

A blazing fast CLI tool that caches AWS resources details locally and provides deep infrastructure tracing

Instead of waiting for slow AWS API calls every time, astat maintains
a local cache for instant querying and visualizes exactly how your
domain requests flow through AWS (DNS -> CloudFront -> LB -> Target -> EC2)

Example workflow:
  $ astat status                	# Check cache status
  $ astat refresh               	# Refresh all services
  $ astat domain trace <domain/uri> # Trace request flow through AWS
  $ astat ec2 list              	# List EC2 instances (instant!)
  $ astat s3 list --refresh     	# Force refresh S3 buckets

Learn more: https://github.com/sunil-saini/astat`,
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if isQuietCommand(cmd) {
			return nil
		}

		if viper.GetBool("auto-refresh") {
			service := ""
			curr := cmd
			for curr.HasParent() {
				if curr.Parent().Name() == "astat" {
					service = curr.Name()
					break
				}
				curr = curr.Parent()
			}

			if service != "" && !isQuietCommand(curr) && service != "domain" {
				switch service {
				case "route53":
					if cmd.Name() == "list" || cmd.Name() == "ls" {
						refresh.AutoRefreshIfStale(cmd.Context(), "route53-zones")
					} else if cmd.Name() == "records" {
						refresh.AutoRefreshIfStale(cmd.Context(), "route53-records")
					}
				case "rds":
					if cmd.Name() == "list" || cmd.Name() == "ls" {
						refresh.AutoRefreshIfStale(cmd.Context(), "rds-clusters")
					} else if cmd.Name() == "instances" {
						refresh.AutoRefreshIfStale(cmd.Context(), "rds-instances")
					}
				default:
					refresh.AutoRefreshIfStale(cmd.Context(), service)
				}
			}
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func isQuietCommand(cmd *cobra.Command) bool {
	return cmd.GroupID == "project" || cmd.Name() == "help"
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil && err != context.Canceled {
		logger.Error("%v", err)
		os.Exit(1)
	}
	refresh.Wait(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/astat/config.yaml)")
	rootCmd.PersistentFlags().String("profile", "", "AWS profile")
	rootCmd.PersistentFlags().String("region", "", "AWS region")
	rootCmd.PersistentFlags().String("output", "table", "output format: table|json")
	rootCmd.PersistentFlags().Bool("refresh", false, "refresh data from AWS")
	rootCmd.PersistentFlags().Duration("ttl", 24*time.Hour, "cache TTL")
	rootCmd.PersistentFlags().Bool("auto-refresh", true, "enable auto refresh if stale")
	rootCmd.PersistentFlags().Int("route53-max-records", 1000, "ignore route53 hosted zones to fetch records with more than max records")

	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("refresh", rootCmd.PersistentFlags().Lookup("refresh"))
	viper.BindPFlag("ttl", rootCmd.PersistentFlags().Lookup("ttl"))
	viper.BindPFlag("auto-refresh", rootCmd.PersistentFlags().Lookup("auto-refresh"))
	viper.BindPFlag("route53-max-records", rootCmd.PersistentFlags().Lookup("route53-max-records"))

	viper.SetDefault("output", "table")
	viper.SetDefault("ttl", 24*time.Hour)
	viper.SetDefault("auto-refresh", true)
	viper.SetDefault("route53-max-records", 1000)

	yellow := color.New(color.FgHiYellow).SprintFunc()

	rootCmd.AddGroup(&cobra.Group{
		ID:    "resources",
		Title: yellow("Resource Commands:"),
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "project",
		Title: yellow("Project Commands:"),
	})

	rootCmd.AddCommand(ec2.EC2Cmd)
	rootCmd.AddCommand(s3.S3Cmd)
	rootCmd.AddCommand(ssm.SSMCmd)
	rootCmd.AddCommand(lambda.LambdaCmd)
	rootCmd.AddCommand(cloudfront.CloudFrontCmd)
	rootCmd.AddCommand(route53.Route53Cmd)
	rootCmd.AddCommand(elb.ElbCmd)
	rootCmd.AddCommand(rds.RDSCmd)
	rootCmd.AddCommand(domain.DomainCmd)
	rootCmd.AddCommand(sqs.SQSCmd)

	rootCmd.AddCommand(ConfigCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upgradeCmd)

	rootCmd.SetUsageTemplate(usageTemplate())
}

func usageTemplate() string {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()

	return fmt.Sprintf(`%s:
  {{.UseLine}}

{{- if .HasAvailableSubCommands}}
{{- range $group := .Groups}}

{{$group.Title}}
{{- range $.Commands}}
{{- if eq .GroupID $group.ID}}
    %s  {{.Short}}
{{- end}}
{{- end}}
{{- end}}

{{- if .HasAvailableSubCommands}}
{{- $ungrouped := false }}
{{- range .Commands}}{{if and (not .GroupID) (or .IsAvailableCommand (eq .Name "help"))}}{{$ungrouped = true}}{{end}}{{end}}
{{- if $ungrouped}}

%s:
{{- range .Commands}}
{{- if and (not .GroupID) (or .IsAvailableCommand (eq .Name "help"))}}
    %s  {{.Short}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}

%s:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{- if .HasAvailableInheritedFlags}}

%s:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{- end}}

{{- if .HasHelpSubCommands}}

%s:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{.CommandPath | printf "%%-14s"}} {{.Short}}{{end}}{{end}}
{{- end}}

{{- if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{- end}}
`,
		yellow("Usage"),
		cyan("{{.Name | printf \"%-12s\"}}"),
		yellow("Other Commands"),
		cyan("{{.Name | printf \"%-12s\"}}"),
		yellow("Flags"),
		yellow("Global Flags"),
		yellow("Additional help topics"),
	)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Failed to get home directory: %v", err)
			return
		}

		viper.AddConfigPath(filepath.Join(home, ".config", "astat"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("ASTAT")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Error("Failed to read config file: %v", err)
		}
	}
}
