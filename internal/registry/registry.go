package registry

import (
	"context"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sunil-saini/astat/internal/aws"
)

type Service struct {
	Name  string
	Fetch func(context.Context, sdkaws.Config) (any, error)
}

var Registry = []Service{
	{
		Name: "ec2",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchEC2Instances(ctx, cfg)
		},
	},
	{
		Name: "s3",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchS3Buckets(ctx, cfg)
		},
	},
	{
		Name: "lambda",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchLambdaFunctions(ctx, cfg)
		},
	},
	{
		Name: "cloudfront",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchCloudFront(ctx, cfg)
		},
	},
	{
		Name: "route53-zones",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchHostedZones(ctx, cfg)
		},
	},
	{
		Name: "route53-records",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchAllRoute53Records(ctx, cfg)
		},
	},
	{
		Name: "ssm",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchSSMParameters(ctx, cfg)
		},
	},
	{
		Name: "elb",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchLoadBalancers(ctx, cfg)
		},
	},
	{
		Name: "rds-clusters",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchRDSClusters(ctx, cfg)
		},
	},
	{
		Name: "rds-instances",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchRDSInstances(ctx, cfg)
		},
	},
	{
		Name: "sqs",
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchSQSQueues(ctx, cfg)
		},
	},
}
