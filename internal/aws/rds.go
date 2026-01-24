package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchRDSInstances(ctx context.Context, cfg aws.Config) ([]model.RDSInstance, error) {
	client := rds.NewFromConfig(cfg)

	roles, err := fetchRDSClusterRoles(ctx, client)
	if err != nil {
		return nil, err
	}

	var instances []model.RDSInstance
	instancePaginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for instancePaginator.HasMorePages() {
		output, err := instancePaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, db := range output.DBInstances {
			instances = append(instances, mapRDSInstance(db, roles))
		}
	}

	return instances, nil
}

func fetchRDSClusterRoles(ctx context.Context, client *rds.Client) (map[string]string, error) {
	roles := make(map[string]string)
	paginator := rds.NewDescribeDBClustersPaginator(client, &rds.DescribeDBClustersInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, cluster := range output.DBClusters {
			extractClusterMemberRoles(cluster.DBClusterMembers, roles)
		}
	}
	return roles, nil
}

func extractClusterMemberRoles(members []rdsTypes.DBClusterMember, roles map[string]string) {
	for _, member := range members {
		if member.DBInstanceIdentifier == nil {
			continue
		}
		role := "Reader"
		if *member.IsClusterWriter {
			role = "Writer"
		}
		roles[*member.DBInstanceIdentifier] = role
	}
}

func mapRDSInstance(db rdsTypes.DBInstance, roles map[string]string) model.RDSInstance {
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

	return model.RDSInstance{
		ClusterIdentifier:  clusterIdentifier,
		InstanceIdentifier: *db.DBInstanceIdentifier,
		Role:               role,
		Engine:             *db.Engine,
		EngineVersion:      *db.EngineVersion,
		DBInstanceStatus:   *db.DBInstanceStatus,
		Endpoint:           endpoint,
		InstanceClass:      *db.DBInstanceClass,
		AvailabilityZone:   *db.AvailabilityZone,
	}
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
			clusters = append(clusters, mapRDSCluster(db))
		}
	}

	return clusters, nil
}

func mapRDSCluster(db rdsTypes.DBCluster) model.RDSCluster {
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

	endpoint := ""
	if db.Endpoint != nil {
		endpoint = *db.Endpoint
	}

	readerEndpoint := ""
	if db.ReaderEndpoint != nil {
		readerEndpoint = *db.ReaderEndpoint
	}

	return model.RDSCluster{
		ClusterIdentifier: *db.DBClusterIdentifier,
		Status:            *db.Status,
		Engine:            *db.Engine,
		EngineVersion:     *db.EngineVersion,
		MultiAZ:           multiAZ,
		InstanceCount:     len(db.DBClusterMembers),
		StorageType:       storageType,
		Endpoint:          endpoint,
		ReaderEndpoint:    readerEndpoint,
		CreateTime:        db.ClusterCreateTime.Format("2006-01-02 15:04:05"),
		IsPublic:          public,
	}
}
