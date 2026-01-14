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
	var dists []model.CloudFrontDistribution
	var marker *string

	for {
		out, err := client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}

		if out.DistributionList == nil || out.DistributionList.IsTruncated == nil || !*out.DistributionList.IsTruncated {
			break
		}

		for _, d := range out.DistributionList.Items {
			origins := make(map[string]string)
			if d.Origins != nil {
				for _, o := range d.Origins.Items {
					origins[*o.Id] = *o.DomainName
				}
			}

			var behaviors []model.CloudFrontBehavior
			if d.CacheBehaviors != nil {
				for _, b := range d.CacheBehaviors.Items {
					behaviors = append(behaviors, model.CloudFrontBehavior{
						PathPattern:    *b.PathPattern,
						TargetOriginID: *b.TargetOriginId,
					})
				}
			}

			defaultOrigin := ""
			if d.DefaultCacheBehavior != nil {
				defaultOrigin = origins[*d.DefaultCacheBehavior.TargetOriginId]
			}

			aliases := ""
			if d.Aliases != nil && len(d.Aliases.Items) > 0 {
				aliases = strings.Join(d.Aliases.Items, ",")
			}

			dists = append(dists, model.CloudFrontDistribution{
				ID:            *d.Id,
				Domain:        *d.DomainName,
				Status:        *d.Status,
				LastUpdated:   d.LastModifiedTime.Format("2006-01-02 15:04:05"),
				Aliases:       aliases,
				Origins:       origins,
				DefaultOrigin: defaultOrigin,
				Behaviors:     behaviors,
			})
		}

		if out.DistributionList.IsTruncated != nil && *out.DistributionList.IsTruncated {
			marker = out.DistributionList.NextMarker
		}
	}

	return dists, nil
}
