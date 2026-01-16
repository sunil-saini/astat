package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/spf13/viper"
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
			zoneType := "public"
			if z.Config.PrivateZone {
				zoneType = "private"
			}

			zones = append(zones, model.Route53HostedZone{
				ID:      *z.Id,
				Name:    *z.Name,
				Type:    zoneType,
				Records: fmt.Sprintf("%d", *z.ResourceRecordSetCount),
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
	maxRecords := viper.GetInt("route53-max-records")

	var allRecords []model.Route53Record
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Limit concurrency to 5

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
			if maxRecords > 0 && zone.ResourceRecordSetCount != nil && int(*zone.ResourceRecordSetCount) > maxRecords {
				continue
			}

			zoneID := strings.TrimPrefix(*zone.Id, "/hostedzone/")
			zoneName := *zone.Name

			wg.Add(1)
			go func(zID, zName string) {
				defer wg.Done()
				sem <- struct{}{}        // Acquire
				defer func() { <-sem }() // Release

				records := fetchZoneRecords(ctx, client, zID, zName)
				mu.Lock()
				allRecords = append(allRecords, records...)
				mu.Unlock()
			}(zoneID, zoneName)
		}

		if !out.IsTruncated {
			break
		}
		marker = out.NextMarker
	}

	wg.Wait()
	return allRecords, nil
}

func fetchZoneRecords(ctx context.Context, client *route53.Client, zoneID, zoneName string) []model.Route53Record {
	var records []model.Route53Record
	var startName *string
	var startType types.RRType

	for {
		if err := ctx.Err(); err != nil {
			return records
		}
		rout, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
			HostedZoneId:    &zoneID,
			StartRecordName: startName,
			StartRecordType: startType,
			MaxItems:        sdkaws.Int32(300),
		})
		if err != nil {
			return records
		}

		for _, r := range rout.ResourceRecordSets {
			records = append(records, mapRoute53RecordSet(zoneName, r))
		}

		if !rout.IsTruncated {
			break
		}
		startName = rout.NextRecordName
		startType = rout.NextRecordType
	}
	return records
}

func mapRoute53RecordSet(zoneName string, r types.ResourceRecordSet) model.Route53Record {
	ttl := "ALIAS"
	if r.TTL != nil {
		ttl = fmt.Sprintf("%d", *r.TTL)
	}

	value := ""
	if r.AliasTarget != nil {
		value = *r.AliasTarget.DNSName
	} else if len(r.ResourceRecords) > 0 {
		value = *r.ResourceRecords[0].Value
	}

	recordType := string(r.Type)
	if r.AliasTarget != nil {
		recordType = "Alias+" + recordType
	}

	return model.Route53Record{
		ZoneName: zoneName,
		Name:     *r.Name,
		Type:     recordType,
		TTL:      ttl,
		Value:    value,
	}
}
