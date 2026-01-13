package aws

import (
	"context"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchCloudFront(ctx context.Context, cfg sdkaws.Config) ([]model.CloudFrontDistribution, error) {
	client := cloudfront.NewFromConfig(cfg)

	out, err := client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{})
	if err != nil {
		return nil, err
	}

	var dists []model.CloudFrontDistribution

	if out.DistributionList == nil {
		return dists, nil
	}

	for _, d := range out.DistributionList.Items {
		origin := ""
		if len(d.Origins.Items) > 0 {
			origin = *d.Origins.Items[0].DomainName
		}

		aliases := ""
		if len(d.Aliases.Items) > 0 {
			aliases = strings.Join(d.Aliases.Items, ",")
		}

		dists = append(dists, model.CloudFrontDistribution{
			ID:          *d.Id,
			Domain:      *d.DomainName,
			Status:      *d.Status,
			Origin:      origin,
			LastUpdated: d.LastModifiedTime.Format("2006-01-02 15:04:05"),
			Aliases:     aliases,
		})
	}

	return dists, nil
}
