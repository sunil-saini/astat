package registry

import (
	"context"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/model"
)

type Service struct {
	Name  string
	Model any
	Fetch func(context.Context, sdkaws.Config) (any, error)
}

var Registry = []Service{
	{
		Name:  "ec2",
		Model: model.EC2Instance{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchEC2Instances(ctx, cfg)
		},
	},
	{
		Name:  "s3",
		Model: model.S3Bucket{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchS3Buckets(ctx, cfg)
		},
	},
	{
		Name:  "lambda",
		Model: model.LambdaFunction{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchLambdaFunctions(ctx, cfg)
		},
	},
	{
		Name:  "cloudfront",
		Model: model.CloudFrontDistribution{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchCloudFront(ctx, cfg)
		},
	},
	{
		Name:  "route53-zones",
		Model: model.Route53HostedZone{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchHostedZones(ctx, cfg)
		},
	},
	{
		Name:  "route53-records",
		Model: model.Route53Record{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchAllRoute53Records(ctx, cfg)
		},
	},
	{
		Name:  "ssm",
		Model: model.SSMParameter{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchSSMParameters(ctx, cfg)
		},
	},
	{
		Name:  "elb",
		Model: model.LoadBalancer{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchLoadBalancers(ctx, cfg)
		},
	},
	{
		Name:  "rds-clusters",
		Model: model.RDSCluster{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchRDSClusters(ctx, cfg)
		},
	},
	{
		Name:  "rds-instances",
		Model: model.RDSInstance{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchRDSInstances(ctx, cfg)
		},
	},
	{
		Name:  "sqs",
		Model: model.SQSQueue{},
		Fetch: func(ctx context.Context, cfg sdkaws.Config) (any, error) {
			return aws.FetchSQSQueues(ctx, cfg)
		},
	},
}
