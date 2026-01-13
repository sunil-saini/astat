package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/cmd/cloudfront"
	"github.com/sunil-saini/astat/cmd/ec2"
	"github.com/sunil-saini/astat/cmd/lambda"
	"github.com/sunil-saini/astat/cmd/route53"
	"github.com/sunil-saini/astat/cmd/s3"
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
	Long: `astat - AWS Status

A blazing fast CLI tool that caches AWS resources details locally,
providing instant access to cloud infrastructure

Instead of waiting for slow AWS API calls every time, astat maintains
a local cache that's automatically refreshed, providing sub millisecond
query times for AWS resources

Example workflow:
  $ astat status              # Check cache status
  $ astat refresh             # Refresh all services
  $ astat ec2 list            # List EC2 instances (instant!)
  $ astat s3 list --refresh   # Force refresh S3 buckets

Learn more: https://github.com/sunil-saini/astat`,
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "status" || cmd.Name() == "help" || cmd.Name() == "config" || cmd.Name() == "completion" || cmd.Name() == "install" || cmd.Name() == "refresh" || cmd.Name() == "version" || cmd.Name() == "upgrade" {
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

			if service != "" && service != "config" && service != "status" && service != "completion" && service != "install" && service != "refresh" && service != "version" && service != "upgrade" {
				if service == "route53" {
					if cmd.Name() == "list" || cmd.Name() == "ls" {
						refresh.AutoRefreshIfStale(cmd.Context(), "route53-zones")
					} else if cmd.Name() == "records" {
						refresh.AutoRefreshIfStale(cmd.Context(), "route53-records")
					}
				} else {
					refresh.AutoRefreshIfStale(cmd.Context(), service)
				}
			}
		}
		return nil
	},
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
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

	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("refresh", rootCmd.PersistentFlags().Lookup("refresh"))
	viper.BindPFlag("ttl", rootCmd.PersistentFlags().Lookup("ttl"))
	viper.BindPFlag("auto-refresh", rootCmd.PersistentFlags().Lookup("auto-refresh"))

	rootCmd.AddCommand(ec2.EC2Cmd)
	rootCmd.AddCommand(s3.S3Cmd)
	rootCmd.AddCommand(ssm.SSMCmd)
	rootCmd.AddCommand(lambda.LambdaCmd)
	rootCmd.AddCommand(cloudfront.CloudFrontCmd)
	rootCmd.AddCommand(route53.Route53Cmd)
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

%s:
  {{.Long}}

%s:
{{- if .HasAvailableSubCommands}}
  {{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
    %s  {{.Short}}{{end}}{{end}}
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
  {{.CommandPath | printf "%-14s"}} {{.Short}}{{end}}{{end}}
{{- end}}

{{- if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{- end}}
`,
		yellow("Usage"),
		yellow("Description"),
		yellow("Available Commands"),
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
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME/.config/astat")
	}

	viper.SetEnvPrefix("CLIDX")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Failed to read config file: %s: %v", viper.ConfigFileUsed(), err)
	}
}
