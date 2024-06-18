package mac

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	awsECS "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

// Idea move away from multi file creation a set outputs as an unified yaml file
// type macdh struct {
// 	ID          string `yaml:"id"`
// 	AZ          string `yaml:"az"`
// 	BackedURL   string `yaml:"backedurl"`
// 	ProjectName string `yaml:"projectname"`
// }

// this creates the stack for the dedicated host
func (r *MacRequest) createDedicatedHost() (dhi *HostInformation, err error) {
	// Get data required for create a dh
	backedURL := getBackedURL()
	r.Region, err = getRegion(r)
	if err != nil {
		return nil, err
	}
	r.AvailabilityZone, err = getAZ(r)
	if err != nil {
		return nil, err
	}
	logging.Debugf("creating a mac %s dedicated host state will be stored at %s",
		r.Architecture, backedURL)
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(stackDedicatedHost),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   backedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: *r.Region}),
		DeployFunc: r.deployerDedicatedHost,
	}
	sr, _ := manager.UpStack(cs)
	dhID, _, err := r.manageResultsDedicatedHost(sr)
	if err != nil {
		return nil, err
	}
	logging.Debugf("mac dedicated host with host id %s has been created successfully", *dhID)
	host, err := data.GetDedicatedHost(*dhID)
	if err != nil {
		return nil, err
	}
	i := getHostInformation(*host)
	dhi = i
	return
}

// this function will create the dedicated host resource
func (r *MacRequest) deployerDedicatedHost(ctx *pulumi.Context) (err error) {
	backedURL := getBackedURL()
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputRegion), pulumi.String(*r.Region))
	dh, err := ec2.NewDedicatedHost(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "dh"),
		&ec2.DedicatedHostArgs{
			AutoPlacement:    pulumi.String("off"),
			AvailabilityZone: pulumi.String(*r.AvailabilityZone),
			InstanceType:     pulumi.String(macTypesByArch[r.Architecture]),
			Tags: maptContext.ResourceTagsWithCustom(
				map[string]string{
					tagKeyBackedURL:         backedURL,
					tagKeyArch:              r.Architecture,
					maptContext.TagKeyRunID: maptContext.RunID(),
				}),
		})
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID), dh.ID())
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostAZ), pulumi.String(*r.AvailabilityZone))
	if err != nil {
		return err
	}
	return nil
}

// results for dedicated host it will return dedicatedhost ID and dedicatedhost AZ
// also write results to files on the target folder
func (r *MacRequest) manageResultsDedicatedHost(stackResult auto.UpResult) (*string, *string, error) {
	if err := output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID): "dedicated_host_id",
	}); err != nil {
		return nil, nil, err
	}
	dhID, ok := stackResult.Outputs[fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID)].Value.(string)
	if !ok {
		return nil, nil, fmt.Errorf("error getting dedicated host ID")
	}
	dhAZ, ok := stackResult.Outputs[fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostAZ)].Value.(string)
	if !ok {
		return nil, nil, fmt.Errorf("error getting dedicated host AZ")
	}
	return &dhID, &dhAZ, nil
}

func (r *MacRequest) ScheduleDestroy(ctx *pulumi.Context) error {
	// https://medium.com/@nilangav/set-up-scheduled-tasks-with-aws-fargate-using-cloudformation-templates-b7bd2f7db46b
	// Cluster is not deleted as it is required to run the self prune container
	clusterName := resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "mac-dh-event-destroy")
	cluster, err := awsECS.NewCluster(ctx,
		clusterName,
		&awsECS.ClusterArgs{
			Tags: maptContext.ResourceTags(),
			Name: pulumi.String(clusterName),
		},
		pulumi.RetainOnDelete(true))
	if err != nil {
		return err
	}

	destroyCmd := []string{"aws", "destroy", "mac"}
	fs, err := ecs.NewFargateService(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "fg"),
		&ecs.FargateServiceArgs{
			Cluster: cluster.Arn,
			TaskDefinitionArgs: &ecs.FargateServiceTaskDefinitionArgs{
				Container: &ecs.TaskDefinitionContainerDefinitionArgs{
					Command: pulumi.ToStringArray(destroyCmd),
					Image:   pulumi.String(""),
				},
			},
		})
	if err != nil {
		return err
	}

	_, err = cloudwatch.NewEventRule(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "mac-dh-event-destroy"),
		&cloudwatch.EventRuleArgs{
			Description: pulumi.String("Destroy event for mac dedicated host"),
			Name: pulumi.String(resourcesUtil.GetResourceName(r.Prefix,
				awsMacMachineID, "mac-dh-event-desotry")),

			// ScheduleExpression: ,
		},
	)
	if err != nil {
		return err
	}

	_, err = cloudwatch.NewEventTarget(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "mac-dh-event-destroy"),
		&cloudwatch.EventTargetArgs{
			EcsTarget: cloudwatch.EventTargetEcsTargetArgs{
				TaskCount:         pulumi.IntPtr(1),
				TaskDefinitionArn: fs.TaskDefinition.Arn(),
			},
		})
	if err != nil {
		return err
	}
	return fmt.Errorf("not implemented yet")
}
