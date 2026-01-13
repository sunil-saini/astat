package aws

import (
	"context"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchS3Buckets(ctx context.Context, cfg sdkaws.Config) ([]model.S3Bucket, error) {
	client := s3.NewFromConfig(cfg)

	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var buckets []model.S3Bucket
	for _, b := range out.Buckets {
		buckets = append(buckets, model.S3Bucket{
			Name:         *b.Name,
			CreationDate: b.CreationDate.Format(time.RFC3339),
		})
	}

	return buckets, nil
}
