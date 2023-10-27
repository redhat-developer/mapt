package mac

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/provider/util/output"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// this creates the stack for the dedicated host
func (r *MacRequest) createDedicatedHost() (*string, *string, error) {
	logging.Debugf("creating dedicated host for mac machine %s", r.Architecture)
	cs := manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName(stackDedicatedHost),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: r.Region}),
		DeployFunc: r.deployerDedicatedHost,
	}
	csResult, err := manager.UpStack(cs)
	if err != nil {
		return nil, nil, err
	}
	dhID, dhAZ, err := r.manageResultsDedicatedHost(
		csResult,
		qenvsContext.GetResultsOutput())
	if err != nil {
		return nil, nil, err
	}
	logging.Debugf("dedicated host with host id %s has been created successfully", *dhID)
	return dhID, dhAZ, nil
}

// this function will create the dedicated host resource
func (r *MacRequest) deployerDedicatedHost(ctx *pulumi.Context) (err error) {
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputRegion), pulumi.String(r.Region))
	az, err := data.GetRandomAvailabilityZone(r.Region)
	if err != nil {
		return
	}
	dh, err := ec2.NewDedicatedHost(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacMachineID, "dh"),
		&ec2.DedicatedHostArgs{
			AutoPlacement:    pulumi.String("off"),
			AvailabilityZone: pulumi.String(*az),
			InstanceType:     pulumi.String(macTypesByArch[r.Architecture]),
			Tags:             qenvsContext.GetTagsAsPulumiStringMap(),
		})
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID), dh.ID())
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostAZ), pulumi.String(*az))
	if err != nil {
		return err
	}
	return nil
}

// results for dedicated host it will return dedicatedhost ID and dedicatedhost AZ
// also write results to files on the target folder
func (r *MacRequest) manageResultsDedicatedHost(stackResult auto.UpResult,
	destinationFolder string) (*string, *string, error) {
	if err := output.Write(stackResult, destinationFolder, map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputDedicatedHostID): "dedicatedHostID",
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
