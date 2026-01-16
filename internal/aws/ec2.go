package aws

import (
	"context"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchEC2Instances(ctx context.Context, cfg sdkaws.Config) ([]model.EC2Instance, error) {
	client := ec2.NewFromConfig(cfg)

	out, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	var instances []model.EC2Instance

	for _, res := range out.Reservations {
		for _, inst := range res.Instances {
			instances = append(instances, mapEC2Instance(inst))
		}
	}

	return instances, nil
}

func mapEC2Instance(inst ec2Types.Instance) model.EC2Instance {
	name := ""
	for _, tag := range inst.Tags {
		if tag.Key != nil && *tag.Key == "Name" {
			name = *tag.Value
			break
		}
	}

	privateIP := ""
	if inst.PrivateIpAddress != nil {
		privateIP = *inst.PrivateIpAddress
	}

	publicIP := ""
	if inst.PublicIpAddress != nil {
		publicIP = *inst.PublicIpAddress
	}

	return model.EC2Instance{
		InstanceID:   *inst.InstanceId,
		Name:         name,
		State:        string(inst.State.Name),
		InstanceType: string(inst.InstanceType),
		AZ:           *inst.Placement.AvailabilityZone,
		PrivateIP:    privateIP,
		PublicIP:     publicIP,
		LaunchTime:   inst.LaunchTime.Format("2006-01-02 15:04:05"),
	}
}
