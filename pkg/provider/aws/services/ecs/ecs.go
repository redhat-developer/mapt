package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
)

func RunTaskWithCommand(region,
	taskDefArn, clusterName,
	containerName, command *string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*region), //
	)
	if err != nil {
		return err
	}
	client := ecs.NewFromConfig(cfg)
	// Run the task
	subnetID, err := data.GetRandomPublicSubnet(*region)
	if err != nil {
		return err
	}
	_, err = client.RunTask(context.TODO(), &ecs.RunTaskInput{
		Cluster:        aws.String(*clusterName),
		TaskDefinition: aws.String(*taskDefArn),
		LaunchType:     types.LaunchTypeFargate,
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        []string{*subnetID},
				AssignPublicIp: types.AssignPublicIpEnabled,
			},
		},
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name:    aws.String(*containerName),
					Command: []string{*command},
				},
			},
		}})
	if err != nil {
		return err
	}
	return nil
}
