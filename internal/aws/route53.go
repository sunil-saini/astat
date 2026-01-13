package aws

import (
	"context"
	"fmt"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/sourcegraph/conc/pool"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchHostedZones(ctx context.Context, cfg sdkaws.Config) ([]model.Route53HostedZone, error) {
	client := route53.NewFromConfig(cfg)

	var zones []model.Route53HostedZone
	var marker *string

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		out, err := client.ListHostedZones(ctx, &route53.ListHostedZonesInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}

		for _, z := range out.HostedZones {
			zones = append(zones, model.Route53HostedZone{
				ID:   *z.Id,
				Name: *z.Name,
			})
		}

		if !out.IsTruncated {
			break
		}
		marker = out.NextMarker
	}

	return zones, nil
}

func FetchAllRoute53Records(ctx context.Context, cfg sdkaws.Config) ([]model.Route53Record, error) {
	client := route53.NewFromConfig(cfg)
	p := pool.NewWithResults[[]model.Route53Record]().
		WithContext(ctx).
		WithMaxGoroutines(5)

	var marker *string
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		out, err := client.ListHostedZones(ctx, &route53.ListHostedZonesInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}

		for _, zone := range out.HostedZones {
			zoneID := strings.TrimPrefix(*zone.Id, "/hostedzone/")
			zoneName := *zone.Name
			p.Go(func(ctx context.Context) ([]model.Route53Record, error) {
				var records []model.Route53Record
				var startName *string
				var startType types.RRType

				for {
					if err := ctx.Err(); err != nil {
						return nil, err
					}
					rout, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
						HostedZoneId:    &zoneID,
						StartRecordName: startName,
						StartRecordType: startType,
						MaxItems:        sdkaws.Int32(300),
					})
					if err != nil {
						return nil, err
					}

					for _, r := range rout.ResourceRecordSets {
						ttl := ""
						if r.TTL != nil {
							ttl = fmt.Sprintf("%d", *r.TTL)
						} else {
							ttl = "ALIAS"
						}

						value := ""
						if r.AliasTarget != nil {
							value = *r.AliasTarget.DNSName
						} else if len(r.ResourceRecords) > 0 {
							value = *r.ResourceRecords[0].Value
						}

						records = append(records, model.Route53Record{
							ZoneName: zoneName,
							Name:     *r.Name,
							Type:     string(r.Type),
							TTL:      ttl,
							Value:    value,
						})
					}

					if !rout.IsTruncated {
						break
					}
					startName = rout.NextRecordName
					startType = rout.NextRecordType
				}
				return records, nil
			})
		}

		if !out.IsTruncated {
			break
		}
		marker = out.NextMarker
	}

	results, err := p.Wait()
	if err != nil {
		return nil, err
	}

	var allRecords []model.Route53Record
	for _, res := range results {
		allRecords = append(allRecords, res...)
	}

	return allRecords, nil
}
