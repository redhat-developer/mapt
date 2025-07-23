package rhelai

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	amiCopy "github.com/redhat-developer/mapt/pkg/provider/aws/modules/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type RHELAIArgs struct {
	Prefix         string
	Version        string
	Arch           string
	ComputeRequest *cr.ComputeRequestArgs
	SubsUsername   string
	SubsUserpass   string
	Spot           bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
}

type rhelAIRequest struct {
	mCtx           *mc.Context
	prefix         *string
	version        *string
	arch           *string
	spot           *bool
	subsUsername   *string
	subsUserpass   *string
	timeout        *string
	allocationData *allocation.AllocationData
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
func Create(mCtxArgs *mc.ContextArgs, args *RHELAIArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := rhelAIRequest{
		mCtx:         mCtx,
		prefix:       &prefix,
		version:      &args.Version,
		arch:         &args.Arch,
		spot:         &args.Spot,
		timeout:      &args.Timeout,
		subsUsername: &args.SubsUsername,
		subsUserpass: &args.SubsUserpass}
	r.allocationData, err = util.IfWithError(args.Spot,
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnSpot(mCtx,
				&args.Prefix, &amiProduct, nil, args.ComputeRequest)
		},
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnDemand()
		})
	if err != nil {
		return err
	}
	amiName := amiName(&args.Version)
	if err = manageAMIReplication(mCtx, &args.Prefix,
		&amiName, r.allocationData.Region); err != nil {
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
		return spot.Destroy(mCtx)
	}
	return nil
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
		fmt.Sprintf("%s*", amiName(r.version)),
		[]string{amiOwnerSelf},
		map[string]string{
			"architecture": amiArch})
	if err != nil {
		return err
	}
	// Networking
	lbEnable := true
	nr := network.NetworkRequest{
		Prefix: *r.prefix,
		ID:     awsRHELDedicatedID,
		Region: *r.allocationData.Region,
		AZ:     *r.allocationData.AZ,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer: &lbEnable,
		Airgap:             false,
	}
	vpc, targetSubnet, _, _, lb, lbEIP, err := nr.Network(ctx, r.mCtx)
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
	securityGroups, err := r.securityGroups(ctx, r.mCtx, vpc)
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		MCtx:           r.mCtx,
		Prefix:         *r.prefix,
		ID:             awsRHELDedicatedID,
		VPC:            vpc,
		Subnet:         targetSubnet,
		AMI:            ami,
		KeyResources:   keyResources,
		SecurityGroups: securityGroups,
		InstaceTypes:   r.allocationData.InstanceTypes,
		DiskSize:       &diskSize,
		LB:             lb,
		LBEIP:          lbEIP,
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
		c.GetHostIP(true))
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

func manageAMIReplication(mCtx *mc.Context, prefix, amiName, region *string) error {
	isAMIOffered, _, err := data.IsAMIOffered(
		data.ImageRequest{
			Name:   amiName,
			Region: region,
			Owner:  &amiOwner})
	if err != nil {
		return err
	}
	if !isAMIOffered {
		acr := amiCopy.CopyAMIRequest{
			MCtx:            mCtx,
			Prefix:          *prefix,
			ID:              awsRHELDedicatedID,
			AMISourceName:   amiName,
			AMISourceArch:   &amiArch,
			AMITargetRegion: region,
			// TODO add this as param
			AMIKeepCopy: true,
		}
		if err := acr.Create(); err != nil {
			return err
		}
	}
	return nil
}
