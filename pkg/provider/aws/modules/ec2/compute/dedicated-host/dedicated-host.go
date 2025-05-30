package dedicatedhost

import (
	"fmt"
	"strings"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilMaps "github.com/redhat-developer/mapt/pkg/util/maps"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	// mapt internal ID for the component: dedicated host
	awsDH = "adh"

	outputDedicatedHostID = "adhDedicatedHostID"
	outputDedicatedHostAZ = "adhDedicatedHostAZ"
	outputRegion          = "adhRegion"
)

var (
	defaultPrefix       = "mapt"
	maptDHDefaultPrefix = "mapt-dh"
)

type DedicatedHostArgs struct {
	Prefix       *string
	Region       *string
	InstanceType *string
	Tags         map[string]string
}

type dedicatedHostRequest struct {
	prefix       *string
	region       *string
	azId         *string
	instanceType *string
	hostId       *string
	tags         map[string]string
}

func (r *dedicatedHostRequest) tagsAsStringMapInput() pulumi.StringMap {
	return utilMaps.Convert(r.tags,
		func(name string) string { return name },
		func(value string) pulumi.StringInput { return pulumi.String(value) })
}

func Create(backedURL *string, exportOutputs bool, args *DedicatedHostArgs) (*ec2Types.Host, error) {
	// here tags contains...
	// all custom + dhTags := utilMaps.Append(maptContext.GetTags(), tags)
	hostId, azId, err := dedicatedHost(args.Region, args.InstanceType, args.Tags)
	if err != nil {
		return nil, err
	}
	r := &dedicatedHostRequest{
		prefix:       util.If(args.Prefix != nil, args.Prefix, &defaultPrefix),
		region:       args.Region,
		azId:         azId,
		instanceType: args.InstanceType,
		hostId:       hostId,
		tags:         args.Tags,
	}
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(maptDHDefaultPrefix),
		ProjectName: fmt.Sprintf("%s-%s", *r.prefix, maptContext.ProjectName()),
		BackedURL:   *backedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION: *args.Region}),
		DeployFunc: r.deploy,
	}
	sr, err := manager.UpStack(cs)
	if err != nil {
		return nil, err
	}
	dhID, _, err := manageResultsDedicatedHost(sr, *r.prefix, exportOutputs)
	if err != nil {
		return nil, err
	}
	logging.Debugf("mac dedicated host with host id %s has been created successfully", *dhID)
	return data.GetDedicatedHost(*dhID)
}

func Destroy(prefix *string) (err error) {
	logging.Debug("Destroy dedicated host resources")
	if prefix == nil {
		prefix = &defaultPrefix
	}
	// Get stack first and export host-id and region
	regionId, dhId, err := getDedicatedHostData(prefix)
	if err != nil {
		return err
	}
	err = aws.DestroyStack(
		aws.DestroyStackRequest{
			ProjectName: fmt.Sprintf("%s-%s",
				*prefix,
				maptContext.ProjectName()),
			Stackname: maptContext.StackNameByProject(maptDHDefaultPrefix),
		})
	if err != nil {
		return err
	}
	return compute.DedicatedHostRelease(regionId, dhId)
}

// Check for a dedicated host across the Azs on a region as we may hit no capacity and there is no way to check for it
// in advance
func dedicatedHost(region, instanceType *string, tags map[string]string) (hostId *string, azId *string, err error) {
	azs := data.GetAvailabilityZones(*region)
	var isOffered bool
	for i := 0; !isOffered && i < len(azs); i++ {
		isOffered, err = data.IsInstanceTypeOfferedByAZ(*region, *instanceType, azs[i])
		if err != nil {
			return
		}
		if !isOffered {
			logging.Debugf("Instancetype %s is not offered at %s", *instanceType, azs[i])
			continue
		}
		hostId, err = compute.DedicatedHost(region, &azs[i], instanceType, tags)
		if err != nil {
			if isCapacityError(err) {
				isOffered = false
				continue
			}
		}
		azId = &azs[i]
		return
	}
	if hostId == nil {
		return nil, nil, fmt.Errorf("no capacity across the region")
	}
	return
}

func isCapacityError(err error) bool {
	return strings.Contains(err.Error(), "Insufficient") ||
		strings.Contains(err.Error(), "capacity")
}

func getDedicatedHostData(prefix *string) (region *string, hostId *string, err error) {
	s, err := manager.CheckStack(
		manager.Stack{
			ProjectName: fmt.Sprintf("%s-%s",
				*prefix,
				maptContext.ProjectName()),
			StackName:           maptContext.StackNameByProject(maptDHDefaultPrefix),
			BackedURL:           maptContext.BackedURL(),
			ProviderCredentials: aws.DefaultCredentials,
		})
	if err != nil {
		return nil, nil, err
	}
	outputs, err := manager.GetOutputs(s)
	if err != nil {
		return nil, nil, err
	}
	if value, exists := outputs[fmt.Sprintf("%s-%s", *prefix, outputRegion)]; exists {
		regionV := value.Value.(string)
		region = &regionV
	}
	if value, exists := outputs[fmt.Sprintf("%s-%s", *prefix, outputDedicatedHostID)]; exists {
		dhIdV := value.Value.(string)
		hostId = &dhIdV
	}
	return
}

// this function will create the dedicated host resource
func (r *dedicatedHostRequest) deploy(ctx *pulumi.Context) (err error) {
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputRegion), pulumi.String(*r.region))
	opts := maptContext.CommonOptions(ctx)
	// Dedicated host will be created through sdk and then
	// imported to be a pulumi managed resource
	// also resource will be destroyed by pulumi
	opts = append(opts, pulumi.Import(pulumi.ID(*r.hostId)))
	dha := &ec2.DedicatedHostArgs{
		AutoPlacement:    pulumi.String("off"),
		AvailabilityZone: pulumi.String(*r.azId),
		InstanceType:     pulumi.String(*r.instanceType),
	}
	// tags need to be exactly the ones added to the dedicated host
	if r.tags != nil {
		dha.Tags = r.tagsAsStringMapInput()
	}
	dh, err := ec2.NewDedicatedHost(ctx,
		resourcesUtil.GetResourceName(*r.prefix, awsDH, "dh"),
		dha, opts...)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputDedicatedHostID),
		dh.ID())
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputDedicatedHostAZ),
		pulumi.String(*r.azId))
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
