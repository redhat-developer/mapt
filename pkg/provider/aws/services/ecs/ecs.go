package ecs

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
)

func RunTaskWithCommand(region,
	taskDefArn, clusterName,
	containerName, command *string,
	subnetID, sgID *string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*region), //
	)
	if err != nil {
		return err
	}
	client := ecs.NewFromConfig(cfg)
	// Run the task
	if subnetID == nil {
		subnetID, err = data.GetRandomPublicSubnet(*region)
		if err != nil {
			return err
		}
	}
	nc := &types.NetworkConfiguration{
		AwsvpcConfiguration: &types.AwsVpcConfiguration{
			Subnets:        []string{*subnetID},
			AssignPublicIp: types.AssignPublicIpEnabled,
		},
	}
	if sgID != nil {
		nc.AwsvpcConfiguration.SecurityGroups = []string{*sgID}
	}
	_, err = client.RunTask(context.TODO(), &ecs.RunTaskInput{
		Cluster:              aws.String(*clusterName),
		TaskDefinition:       aws.String(*taskDefArn),
		LaunchType:           types.LaunchTypeFargate,
		NetworkConfiguration: nc,
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name:    aws.String(*containerName),
					Command: strings.Fields(*command),
				},
			},
		}})
	if err != nil {
		return err
	}
	return nil
}

func GetTags(region, taskDefArn *string) (map[string]*string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*region), //
	)
	if err != nil {
		return nil, err
	}
	client := ecs.NewFromConfig(cfg)
	out, err := client.ListTagsForResource(context.TODO(),
		&ecs.ListTagsForResourceInput{
			ResourceArn: aws.String(*taskDefArn),
		})
	if err != nil {
		return nil, err
	}
	var tags = make(map[string]*string)
	for _, tag := range out.Tags {
		tags[*tag.Key] = tag.Value
	}
	return tags, nil
}
