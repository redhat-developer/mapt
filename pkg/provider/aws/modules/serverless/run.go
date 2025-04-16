package serverless

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/scheduler"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type RunServerlessTaskArgs struct {
	Region        string
	TaskArn       string
	ContainerName string
	Command       string
	// Optional to identify in case orchestrate several runs
	Prefix      string
	ComponentID string
}

func RunServerlessTask(ctx *pulumi.Context, args *RunServerlessTaskArgs) error {
	clusterArn, err := data.GetCluster(
		MaptServerlessClusterName,
		args.Region)
	if err != nil {
		return err
	}
	sRoleArn, err := data.GetRole(maptServerlessExecRoleName)
	if err != nil {
		return err
	}
	rn := resourcesUtil.GetResourceName(
		util.If(len(args.Prefix) == 0, defaultPrefix, args.Prefix),
		util.If(len(args.ComponentID) == 0, defaultComponentID, args.ComponentID),
		"fgs")
	// Define ECS Task as Target with Command Override
	// targets := []eventbridge. Target{
	// 	{
	// 		Id:      aws.String("1"),
	// 		Arn:     aws.String(clusterArn),
	// 		RoleArn: aws.String(roleArn),
	// 		EcsParameters: &eventbridge.EcsParameters{
	// 			TaskDefinitionArn: aws.String(taskArn),
	// 			LaunchType:        "FARGATE",
	// 			NetworkConfiguration: &eventbridge.NetworkConfiguration{
	// 				AwsvpcConfiguration: &eventbridge.AwsVpcConfiguration{
	// 					Subnets:        []string{"subnet-abc123", "subnet-def456"},
	// 					SecurityGroups: []string{"sg-xyz789"},
	// 					AssignPublicIp: "ENABLED",
	// 				},
	// 			},
	// 			Overrides: &eventbridge.EcsTaskOverride{
	// 				ContainerOverrides: []eventbridge.EcsContainerOverride{
	// 					{
	// 						Name:    aws.String(containerName),
	// 						Command: []string{"your-new-command", "arg1", "arg2"},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// // Convert targets to JSON
	// targetsJSON, err := json.Marshal(targets)
	// if err != nil {
	// 	log.Fatalf("Failed to marshal targets: %v", err)
	// }
	subnetID, err := data.GetRandomPublicSubnet(args.Region)
	if err != nil {
		return err
	}
	ses, err := generateOneTimeScheduleExpression(args.Region, "30s")
	if err != nil {
		return err
	}
	se := scheduleExpressionByType(&OneTime, ses)
	_, err = scheduler.NewSchedule(ctx,
		rn,
		&scheduler.ScheduleArgs{
			FlexibleTimeWindow: scheduler.ScheduleFlexibleTimeWindowArgs{
				Mode:                   scheduler.ScheduleFlexibleTimeWindowModeFlexible,
				MaximumWindowInMinutes: pulumi.Float64(1),
			},
			Target: scheduler.ScheduleTargetArgs{
				EcsParameters: scheduler.ScheduleEcsParametersArgs{
					TaskDefinitionArn: pulumi.String(args.TaskArn),
					LaunchType:        scheduler.ScheduleLaunchTypeFargate,
					NetworkConfiguration: scheduler.ScheduleNetworkConfigurationArgs{
						// https://github.com/aws/aws-cdk/issues/13348#issuecomment-1539336376
						AwsvpcConfiguration: scheduler.ScheduleAwsVpcConfigurationArgs{
							AssignPublicIp: scheduler.ScheduleAssignPublicIpEnabled,
							Subnets: pulumi.StringArray{
								pulumi.String(*subnetID),
							},
						},
					},
				},
				Arn:     pulumi.String(*clusterArn),
				RoleArn: pulumi.String(*sRoleArn),
			},
			ScheduleExpression:         pulumi.String(*se),
			ScheduleExpressionTimezone: pulumi.String(data.RegionTimezones[args.Region]),
		})
	if err != nil {
		return err
	}
	return fmt.Errorf("not implemented yet")
}
