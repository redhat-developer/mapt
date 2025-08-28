package ami

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
)

type CopyAMIRequest struct {
	MCtx          *mc.Context `validate:"required"`
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

func (r *CopyAMIRequest) validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

// Create will check if ami copy state exists and if not it will create the stack
func (r CopyAMIRequest) Create() error {
	if err := r.validate(); err != nil {
		return err
	}
	_, err := manager.CheckStack(manager.Stack{
		StackName:   r.MCtx.StackNameByProject("copyAMI"),
		ProjectName: r.MCtx.ProjectName(),
		BackedURL:   r.MCtx.BackedURL()})
	if err != nil {
		return r.createStack(r.MCtx)
	}
	return nil
}

// Check if spot option stack was created on the backed url
func Exist(mCtx *mc.Context) bool {
	s, err := manager.CheckStack(manager.Stack{
		StackName:   mCtx.StackNameByProject("copyAMI"),
		ProjectName: mCtx.ProjectName(),
		BackedURL:   mCtx.BackedURL()})
	return err == nil && s != nil
}

// Destroy the stack
func Destroy(mCtx *mc.Context) (err error) {
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject("copyAMI"),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials}
	return manager.DestroyStack(mCtx, stack)
}

// function to create the stack
func (r CopyAMIRequest) createStack(mCtx *mc.Context) error {
	credentials := aws.DefaultCredentials
	if r.AMITargetRegion != nil {
		credentials = aws.GetClouProviderCredentials(map[string]string{
			awsConstants.CONFIG_AWS_REGION: *r.AMITargetRegion,
		})
	}
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject("copyAMI"),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: credentials,
		DeployFunc:          r.deployer,
	}
	_, err := manager.UpStack(mCtx, stack)
	return err
}

// deployer function to create the logic to get the best spot option
// and it will export the data from the best spot option to the stack state
func (r CopyAMIRequest) deployer(ctx *pulumi.Context) error {
	// find were the ami is
	amiInfo, err := data.FindAMI(r.AMISourceName, r.AMISourceArch)
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
				Tags: r.MCtx.ResourceTagsWithCustom(
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
