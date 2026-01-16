package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchRDSInstances(ctx context.Context, cfg aws.Config) ([]model.RDSInstance, error) {
	client := rds.NewFromConfig(cfg)

	// Map to store roles: InstanceIdentifier -> Role
	roles := make(map[string]string)

	// Fetch clusters to identify writer/reader roles
	clusterPaginator := rds.NewDescribeDBClustersPaginator(client, &rds.DescribeDBClustersInput{})
	for clusterPaginator.HasMorePages() {
		output, err := clusterPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, cluster := range output.DBClusters {
			for _, member := range cluster.DBClusterMembers {
				role := "Reader"
				if *member.IsClusterWriter {
					role = "Writer"
				}
				if member.DBInstanceIdentifier != nil {
					roles[*member.DBInstanceIdentifier] = role
				}
			}
		}
	}

	var instances []model.RDSInstance
	instancePaginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for instancePaginator.HasMorePages() {
		output, err := instancePaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, db := range output.DBInstances {
			endpoint := ""
			if db.Endpoint != nil {
				endpoint = *db.Endpoint.Address
			}

			clusterIdentifier := "Standalone"
			if db.DBClusterIdentifier != nil {
				clusterIdentifier = *db.DBClusterIdentifier
			}

			role, ok := roles[*db.DBInstanceIdentifier]
			if !ok {
				role = "Writer" // Standalone are Writers
			}

			instances = append(instances, model.RDSInstance{
				ClusterIdentifier:  clusterIdentifier,
				InstanceIdentifier: *db.DBInstanceIdentifier,
				Role:               role,
				Engine:             *db.Engine,
				EngineVersion:      *db.EngineVersion,
				DBInstanceStatus:   *db.DBInstanceStatus,
				Endpoint:           endpoint,
				InstanceClass:      *db.DBInstanceClass,
				AvailabilityZone:   *db.AvailabilityZone,
			})
		}
	}

	return instances, nil
}

func FetchRDSClusters(ctx context.Context, cfg aws.Config) ([]model.RDSCluster, error) {
	client := rds.NewFromConfig(cfg)
	var clusters []model.RDSCluster

	clusterPaginator := rds.NewDescribeDBClustersPaginator(client, &rds.DescribeDBClustersInput{})
	for clusterPaginator.HasMorePages() {
		output, err := clusterPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, db := range output.DBClusters {
			storageType := ""
			if db.StorageType != nil {
				storageType = *db.StorageType
			}

			multiAZ := "No"
			if db.MultiAZ != nil && *db.MultiAZ {
				multiAZ = "Yes"
			}

			public := "No"
			if db.PubliclyAccessible != nil && *db.PubliclyAccessible {
				public = "Yes"
			}

			clusters = append(clusters, model.RDSCluster{
				ClusterIdentifier: *db.DBClusterIdentifier,
				Status:            *db.Status,
				Engine:            *db.Engine,
				EngineVersion:     *db.EngineVersion,
				MultiAZ:           multiAZ,
				InstanceCount:     len(db.DBClusterMembers),
				StorageType:       storageType,
				CreateTime:        db.ClusterCreateTime.Format("2006-01-02 15:04:05"),
				IsPublic:          public,
			})
		}
	}

	return clusters, nil
}
