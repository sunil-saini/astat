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

	paginator := s3.NewListBucketsPaginator(client, &s3.ListBucketsInput{
		MaxBuckets: sdkaws.Int32(1000),
	})

	var buckets []model.S3Bucket
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, b := range page.Buckets {
			buckets = append(buckets, model.S3Bucket{
				Name:         *b.Name,
				CreationDate: b.CreationDate.Format(time.RFC3339),
				Region:       sdkaws.ToString(b.BucketRegion),
			})
		}
	}

	return buckets, nil
}
