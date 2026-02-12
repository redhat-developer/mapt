package rhelai

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	apiRHELAI "github.com/redhat-developer/mapt/pkg/target/host/rhelai"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type rhelAIRequest struct {
	mCtx           *mc.Context
	prefix         *string
	amiName        *string
	arch           *string
	spot           bool
	timeout        *string
	allocationData *allocation.AllocationResult
}

func (r *rhelAIRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

// Create orchestrate 2 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(mCtxArgs *mc.ContextArgs, args *apiRHELAI.RHELAIArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	// Compose request
	amiName := amiName(&args.Accelerator, &args.Version)
	if len(args.CustomAMI) != 0 {
		amiName = fmt.Sprintf("%s*", args.CustomAMI)
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := rhelAIRequest{
		mCtx:    mCtx,
		prefix:  &prefix,
		amiName: &amiName,
		arch:    &args.Arch,
		timeout: &args.Timeout}
	if args.Spot != nil {
		r.spot = args.Spot.Spot
	}
	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			Prefix:                &args.Prefix,
			ComputeRequest:        args.ComputeRequest,
			AMIProductDescription: &amiProduct,
			Spot:                  args.Spot,
		})
	if err != nil {
		return err
	}
	if err = checkAMIExists(mCtx.Context(), &amiName, r.allocationData.Region, &amiArch); err != nil {
		return err
	}
	return r.createMachine()
}

// Will destroy resources related to machine
func Destroy(mCtxArgs *mc.ContextArgs) error {
	logging.Debug("Run rhel destroy")
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}

	if err := aws.DestroyStack(
		mCtx,
		aws.DestroyStackRequest{
			Stackname: stackName,
		}); err != nil {
		return err
	}
	if spot.Exist(mCtx) {
		if err := spot.Destroy(mCtx); err != nil {
			return err
		}
	}

	// Cleanup S3 state after all stacks have been destroyed
	return aws.CleanupState(mCtx)
}

func (r *rhelAIRequest) createMachine() error {
	cs := manager.Stack{
		StackName:   r.mCtx.StackNameByProject(stackName),
		ProjectName: r.mCtx.ProjectName(),
		BackedURL:   r.mCtx.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION: *r.allocationData.Region}),
		DeployFunc: r.deploy,
	}

	sr, _ := manager.UpStack(r.mCtx, cs)
	return r.manageResults(sr)
}

// function wil all the logic to deploy resources required by windows
// * create AMI Copy if needed
// * networking
// * key
// * security group
// * compute
// * checks
func (r *rhelAIRequest) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		*r.amiName,
		[]string{amiOwner},
		map[string]string{
			"architecture": amiArch})
	if err != nil {
		return err
	}
	// Networking
	nw, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:             *r.prefix,
			ID:                 awsRHELDedicatedID,
			Region:             *r.allocationData.Region,
			AZ:                 *r.allocationData.AZ,
			CreateLoadBalancer: true,
		})
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, awsRHELDedicatedID, "pk")}
	keyResources, err := kpr.Create(ctx, r.mCtx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	// Security groups
	securityGroups, err := r.securityGroups(ctx, r.mCtx, nw.Vpc)
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		MCtx:           r.mCtx,
		Prefix:         *r.prefix,
		ID:             awsRHELDedicatedID,
		VPC:            nw.Vpc,
		Subnet:         nw.Subnet,
		AMI:            ami,
		KeyResources:   keyResources,
		SecurityGroups: securityGroups,
		InstaceTypes:   r.allocationData.InstanceTypes,
		DiskSize:       &diskSize,
		LB:             nw.LoadBalancer,
		Eip:            nw.Eip,
		LBTargetGroups: []int{22}}
	if r.allocationData.SpotPrice != nil {
		cr.Spot = true
		cr.SpotPrice = *r.allocationData.SpotPrice
	}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername),
		pulumi.String(amiUserDefault))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost),
		c.GetHostDnsName(true))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx, r.mCtx,
			*r.allocationData.Region, *r.prefix,
			awsRHELDedicatedID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"rhel",
				r.mCtx.ProjectName(),
				r.mCtx.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	return c.Readiness(ctx, command.CommandPing, *r.prefix, awsRHELDedicatedID,
		keyResources.PrivateKey, amiUserDefault, nil, c.Dependencies)
}

// Write exported values in context to files o a selected target folder
func (r *rhelAIRequest) manageResults(stackResult auto.UpResult) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *r.prefix, outputHost):           "host",
	}
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), results)
}

// security group for mac machine with ingress rules for ssh and vnc
func (r *rhelAIRequest) securityGroups(ctx *pulumi.Context, mCtx *mc.Context,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// ingress for ssh access from 0.0.0.0
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(*r.prefix, awsRHELDedicatedID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsRHELDedicatedID),
		IngressRules: []securityGroup.IngressRules{
			sshIngressRule},
	}.Create(ctx, mCtx)
	if err != nil {
		return nil, err
	}
	// Convert to an array of IDs
	sgs := util.ArrayConvert([]*ec2.SecurityGroup{sg.SG},
		func(sg *ec2.SecurityGroup) pulumi.StringInput {
			return sg.ID()
		})
	return pulumi.StringArray(sgs[:]), nil
}

func checkAMIExists(ctx context.Context, amiName, region, arch *string) error {
	isAMIOffered, _, err := data.IsAMIOffered(
		ctx,
		data.ImageRequest{
			Name:   amiName,
			Arch:   arch,
			Region: region,
			Owner:  &amiOwner})
	if err != nil {
		return err
	}
	if !isAMIOffered {
		return fmt.Errorf("AMI %s could not be found in region: %s", *amiName, *region)
	}
	return nil
}
