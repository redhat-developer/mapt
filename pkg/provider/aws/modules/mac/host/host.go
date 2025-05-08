package host

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

func CreatePoolDedicatedHost(args *PoolMacDedicatedHostRequestArgs) (dhi *mac.HostInformation, err error) {
	tags := map[string]string{
		macConstants.TagKeyBackedURL: args.BackedURL,
		macConstants.TagKeyPrefix:    args.MacDedicatedHost.Prefix,
		maptContext.TagKeyRunID:      maptContext.RunID(),
	}
	maps.Copy(tags, args.PoolID.AsTags())
	return createDedicatedHost(args.MacDedicatedHost, args.BackedURL, tags, false)
}

// this creates the stack for the dedicated host
func CreateDedicatedHost(args *MacDedicatedHostRequestArgs) (dhi *mac.HostInformation, err error) {
	backedURL := getBackedURL()
	tags := map[string]string{
		macConstants.TagKeyBackedURL: backedURL,
		macConstants.TagKeyPrefix:    args.Prefix,
		macConstants.TagKeyArch:      args.Architecture,
		maptContext.TagKeyRunID:      maptContext.RunID(),
		macConstants.TagKeyTicket:    "",
	}
	return createDedicatedHost(args, backedURL, tags, true)
}

func createDedicatedHost(args *MacDedicatedHostRequestArgs,
	backedURL string,
	tags map[string]string,
	exportOutputs bool) (dhi *mac.HostInformation, err error) {
	// mac does not offer spot, this will check region based on default region Env,
	// if machine type is not offered on the region it will try to find a new region for it
	dHArgs := dedicatedHostArgs{
		prefix: args.Prefix,
		arch:   args.Architecture,
		tags:   tags,
	}
	dHArgs.region, err = getRegion(args.Architecture, true)
	if err != nil {
		return nil, err
	}
	// We will try on each Az in case we do not have capacity
	sr, err := retryCreateStack(&dHArgs, &backedURL)
	if err != nil {
		return nil, err
	}
	dhID, _, err := manageResultsDedicatedHost(sr, dHArgs.prefix, exportOutputs)
	if err != nil {
		return nil, err
	}
	logging.Debugf("mac dedicated host with host id %s has been created successfully", *dhID)
	host, err := data.GetDedicatedHost(*dhID)
	if err != nil {
		return nil, err
	}
	i := GetHostInformation(*host)
	dhi = i
	return
}

func retryCreateStack(dHArgs *dedicatedHostArgs, backedURL *string) (sr auto.UpResult, err error) {
	created := false
	azs := data.GetAvailabilityZones(*dHArgs.region)
	for i := 0; created || i < len(azs); i++ {
		// for _, az := range data.GetAvailabilityZones(*dHArgs.region) {
		dHArgs.availabilityZone = &azs[i]
		cs := manager.Stack{
			StackName:   maptContext.StackNameByProject(mac.StackDedicatedHost),
			ProjectName: maptContext.ProjectName(),
			BackedURL:   *backedURL,
			ProviderCredentials: aws.GetClouProviderCredentials(
				map[string]string{
					awsConstants.CONFIG_AWS_REGION: *dHArgs.region}),
			DeployFunc: dHArgs.deploy,
		}
		sr, err = manager.UpStack(cs)
		if err != nil {
			if isCapacityError(err) {
				break
			}
			return
		}
		created = true
	}
	if !created {
		err = fmt.Errorf("currently no AZ on %s has capacity", *dHArgs.region)
	}
	return
}

func isCapacityError(err error) bool {
	return strings.Contains(err.Error(), "Insufficient") ||
		strings.Contains(err.Error(), "capacity")
}

// this function will create the dedicated host resource
func (r *dedicatedHostArgs) deploy(ctx *pulumi.Context) (err error) {
	ctx.Export(fmt.Sprintf("%s-%s", r.prefix, outputRegion), pulumi.String(*r.region))
	dh, err := ec2.NewDedicatedHost(ctx,
		resourcesUtil.GetResourceName(r.prefix, awsMacHostID, "dh"),
		&ec2.DedicatedHostArgs{
			AutoPlacement:    pulumi.String("off"),
			AvailabilityZone: pulumi.String(*r.availabilityZone),
			InstanceType:     pulumi.String(mac.TypesByArch[r.arch]),
			Tags:             maptContext.ResourceTagsWithCustom(r.tags),
		}, maptContext.CommonOptions(ctx)...)
	ctx.Export(fmt.Sprintf("%s-%s", r.prefix, outputDedicatedHostID),
		dh.ID())
	ctx.Export(fmt.Sprintf("%s-%s", r.prefix, outputDedicatedHostAZ),
		pulumi.String(*r.availabilityZone))
	if err != nil {
		return err
	}
	return nil
}

// results for dedicated host it will return dedicatedhost ID and dedicatedhost AZ
// also write results to files on the target folder
func manageResultsDedicatedHost(
	stackResult auto.UpResult, prefix string, export bool) (*string, *string, error) {
	if export {
		if err := output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
			fmt.Sprintf("%s-%s", prefix, outputDedicatedHostID): "dedicated_host_id",
		}); err != nil {
			return nil, nil, err
		}
	}
	dhID, ok := stackResult.Outputs[fmt.Sprintf("%s-%s", prefix, outputDedicatedHostID)].Value.(string)
	if !ok {
		return nil, nil, fmt.Errorf("error getting dedicated host ID")
	}
	dhAZ, ok := stackResult.Outputs[fmt.Sprintf("%s-%s", prefix, outputDedicatedHostAZ)].Value.(string)
	if !ok {
		return nil, nil, fmt.Errorf("error getting dedicated host AZ")
	}
	return &dhID, &dhAZ, nil
}
