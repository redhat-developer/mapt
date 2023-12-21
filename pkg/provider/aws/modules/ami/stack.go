package ami

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	amiSVC "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/ami"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CopyAMIRequest struct {
	Prefix        string
	ID            string
	AMISourceName *string
	AMISourceArch *string
	// AMITargetName string
	// If AMITargetRegion is nil
	// it will be copied to all regions
	AMITargetRegion *string
	// if set to true it will keep the AMI on destroy
	AMIKeepCopy bool
	// Only avai for windows images, this will create fast laucn
	FastLaunch  bool
	MaxParallel int32 // number of snapshost to support fast enable
}

// Create wil get the information for the best spot choice it is backed
// within a stack and state to allow idempotency, otherwise run 2nd time a create
// may bring other region with best option and then all dependant resources from other
// stacks would need to be updated

// So create will check if stack with state already exists, if that is the case it will
// pick info from its outputs
// If stack does not exists it will create it
func (r CopyAMIRequest) Create() error {
	_, err := manager.CheckStack(manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName("copyAMI"),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL()})
	if err != nil {
		return r.createStack()
	}
	return nil
}

// Check if spot option stack was created on the backed url
func Exist() bool {
	s, err := manager.CheckStack(manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName("copyAMI"),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL()})
	return err == nil && s != nil
}

// Destroy the stack
func Destroy() (err error) {
	stack := manager.Stack{
		StackName:           qenvsContext.GetStackInstanceName("copyAMI"),
		ProjectName:         qenvsContext.GetInstanceName(),
		BackedURL:           qenvsContext.GetBackedURL(),
		ProviderCredentials: aws.DefaultCredentials}
	return manager.DestroyStack(stack)
}

// function to create the stack
func (r CopyAMIRequest) createStack() error {
	credentials := aws.DefaultCredentials
	if r.AMITargetRegion != nil {
		credentials = aws.GetClouProviderCredentials(map[string]string{
			aws.CONFIG_AWS_REGION: *r.AMITargetRegion,
		})
	}
	stack := manager.Stack{
		StackName:           qenvsContext.GetStackInstanceName("copyAMI"),
		ProjectName:         qenvsContext.GetInstanceName(),
		BackedURL:           qenvsContext.GetBackedURL(),
		ProviderCredentials: credentials,
		DeployFunc:          r.deployer,
	}
	_, err := manager.UpStack(stack)
	return err
}

// deployer function to create the logic to get the best spot option
// and it will export the data from the best spot option to the stack state
func (r CopyAMIRequest) deployer(ctx *pulumi.Context) error {
	// find were the ami is
	amiInfo, err := amiSVC.FindAMI(r.AMISourceName, r.AMISourceArch)
	if err != nil {
		return err
	}
	// if target region is the same region as the source we do not copy
	// mostly for covering use case on copy to all regions (all except the source)
	if amiInfo.Region != r.AMITargetRegion {
		ami, err := ec2.NewAmiCopy(ctx,
			*r.AMISourceName,
			&ec2.AmiCopyArgs{
				Description: pulumi.String(
					fmt.Sprintf("Replica of %s from %s", *amiInfo.Image.ImageId, *amiInfo.Region)),
				SourceAmiId:     pulumi.String(*amiInfo.Image.ImageId),
				SourceAmiRegion: pulumi.String(*amiInfo.Region),
				Tags: qenvsContext.ResourceTagsWithCustom(
					map[string]string{"Name": *r.AMISourceName}),
			},
			pulumi.RetainOnDelete(r.AMIKeepCopy))
		if err != nil {
			return err
		}
		if r.FastLaunch {
			_ = ami.ID().ApplyT(func(amiID string) error {
				return amiSVC.EnableFastLaunch(
					r.AMITargetRegion,
					&amiID,
					&r.MaxParallel)
			})
		}
	}
	return nil
}
