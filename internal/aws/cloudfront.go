package aws

import (
	"context"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cfTypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchCloudFront(ctx context.Context, cfg sdkaws.Config) ([]model.CloudFrontDistribution, error) {
	client := cloudfront.NewFromConfig(cfg)

	// Fetch all tenants to map domains for multi-tenant distributions
	tenantsByDist, _ := fetchAllTenants(ctx, client)

	var dists []model.CloudFrontDistribution
	var marker *string

	for {
		out, err := client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}

		if out.DistributionList != nil {
			for _, d := range out.DistributionList.Items {
				dists = append(dists, mapCloudFrontDistribution(d, tenantsByDist))
			}
		}

		if out.DistributionList == nil || out.DistributionList.IsTruncated == nil || !*out.DistributionList.IsTruncated {
			break
		}
		marker = out.DistributionList.NextMarker
	}

	return dists, nil
}

func mapCloudFrontDistribution(d cfTypes.DistributionSummary, tenantsByDist map[string][]string) model.CloudFrontDistribution {
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
	aliasList := d.Aliases.Items
	if tenantDomains, ok := tenantsByDist[*d.Id]; ok {
		aliasList = append(aliasList, tenantDomains...)
	}

	if len(aliasList) > 0 {
		aliases = "- " + strings.Join(aliasList, "\n- ")
	}

	distType := "Standard"
	if d.ConnectionMode == cfTypes.ConnectionModeTenantOnly {
		distType = "Multi-tenant"
	}

	return model.CloudFrontDistribution{
		ID:            *d.Id,
		Domain:        *d.DomainName,
		Status:        *d.Status,
		Type:          distType,
		LastUpdated:   d.LastModifiedTime.Format("2006-01-02 15:04:05"),
		Aliases:       aliases,
		Origins:       origins,
		DefaultOrigin: defaultOrigin,
		Behaviors:     behaviors,
	}
}

func fetchAllTenants(ctx context.Context, client *cloudfront.Client) (map[string][]string, error) {
	tenantsByDist := make(map[string][]string)
	var marker *string

	for {
		out, err := client.ListDistributionTenants(ctx, &cloudfront.ListDistributionTenantsInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}

		for _, t := range out.DistributionTenantList {
			addTenantDomains(t, tenantsByDist)
		}

		if out.NextMarker == nil || *out.NextMarker == "" {
			break
		}
		marker = out.NextMarker
	}

	return tenantsByDist, nil
}

func addTenantDomains(t cfTypes.DistributionTenantSummary, tenantsByDist map[string][]string) {
	if t.DistributionId == nil {
		return
	}

	distID := *t.DistributionId
	var domains []string
	for _, d := range t.Domains {
		if d.Domain != nil {
			domains = append(domains, *d.Domain)
		}
	}

	if len(domains) > 0 {
		tenantsByDist[distID] = append(tenantsByDist[distID], domains...)
	}
}
