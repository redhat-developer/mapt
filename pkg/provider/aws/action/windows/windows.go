package windows

import (
	_ "embed"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	amiCopy "github.com/redhat-developer/mapt/pkg/provider/aws/modules/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	cloudConfigWindowsServer "github.com/redhat-developer/mapt/pkg/provider/util/cloud-config/windows-server"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

// add proxy https://github.com/ptcodes/proxy-server-with-terraform/blob/master/main.tf
type WindowsServerArgs struct {
	Prefix string
	// AMI info. Optional. User and Owner only applied
	// if AMIName is set
	AMIName     string
	AMIUser     string
	AMIOwner    string
	AMILang     string
	AMIKeepCopy bool
	// Machine params
	ComputeRequest *cr.ComputeRequestArgs
	Spot           *spotTypes.SpotArgs
	Airgap         bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
}

type windowsServerRequest struct {
	mCtx   *mc.Context
	prefix *string

	amiName     *string
	amiUser     *string
	amiOwner    *string
	amiLang     *string
	amiKeepCopy *bool

	spot           bool
	timeout        *string
	allocationData *allocation.AllocationResult
	airgap         *bool
	// internal management
	// For airgap scenario there is an orchestation of
	// a phase with connectivity on the machine (allowing bootstraping)
	// a pahase with connectivyt off where the subnet for the target lost the nat gateway
	airgapPhaseConnectivity network.Connectivity
}

func (r *windowsServerRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(mCtxArgs *mc.ContextArgs, args *WindowsServerArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	if len(args.AMIName) == 0 {
		args.AMIName = amiNameDefault
		args.AMIUser = amiUserDefault
		args.AMIOwner = amiOwnerDefault
	}
	if len(args.AMILang) > 0 && args.AMILang == amiLangNonEng {
		args.AMIName = amiNonEngNameDefault
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := windowsServerRequest{
		mCtx:        mCtx,
		prefix:      &prefix,
		amiName:     &args.AMIName,
		amiUser:     &args.AMIUser,
		amiOwner:    &args.AMIOwner,
		amiKeepCopy: &args.AMIKeepCopy,
		amiLang:     &args.AMILang,
		timeout:     &args.Timeout,
		airgap:      &args.Airgap}
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
	isAMIOffered, _, err := data.IsAMIOffered(
		mCtx.Context(),
		data.ImageRequest{
			Name:   r.amiName,
			Region: r.allocationData.Region})
	if err != nil {
		return err
	}
	// If it is not offered need to create a copy on the target region
	if !isAMIOffered {
		acr := amiCopy.CopyAMIRequest{
			MCtx:            mCtx,
			Prefix:          *r.prefix,
			ID:              awsWindowsDedicatedID,
			AMISourceName:   r.amiName,
			AMISourceArch:   nil,
			AMITargetRegion: r.allocationData.Region,
			AMIKeepCopy:     *r.amiKeepCopy,
			FastLaunch:      amiFastLaunch,
			MaxParallel:     int32(amiFastLaunchMaxParallel),
		}
		if err := acr.Create(); err != nil {
			return err
		}
	}
	// if not only host the mac machine will be created
	if !*r.airgap {
		return r.createMachine()
	}
	// Airgap scneario requires orchestration
	return r.createAirgapMachine()
}

// Will destroy resources related to machine
func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	logging.Debug("Run windows destroy")
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
	if amiCopy.Exist(mCtx) {
		err = amiCopy.Destroy(mCtx)
		if err != nil {
			return
		}
	}
	if spot.Exist(mCtx) {
		if err := spot.Destroy(mCtx); err != nil {
			return err
		}
	}

	// Cleanup S3 state after all stacks have been destroyed
	return aws.CleanupState(mCtx)
}

func (r *windowsServerRequest) createMachine() error {
	cs := manager.Stack{
		StackName:   r.mCtx.StackNameByProject(stackName),
		ProjectName: r.mCtx.ProjectName(),
		BackedURL:   r.mCtx.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION: *r.allocationData.Region}),
		DeployFunc: r.deploy,
	}

	sr, err := manager.UpStack(r.mCtx, cs)
	if err != nil {
		return err
	}
	return manageResults(r.mCtx, sr, r.prefix, r.airgap)
}

// Abstract this with a stackAirgapHandle receives a fn (connectivty on / off) err executes
// first on then off
func (r *windowsServerRequest) createAirgapMachine() error {
	r.airgapPhaseConnectivity = network.ON
	err := r.createMachine()
	if err != nil {
		return nil
	}
	r.airgapPhaseConnectivity = network.OFF
	return r.createMachine()
}

// function wil all the logic to deploy resources required by windows
// * create AMI Copy if needed
// * networking
// * key
// * security group
// * compute
// * checks
func (r *windowsServerRequest) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get AMI ref
	// ami, err := amiSVC.GetAMIByName(ctx, r.AMIName, r.AMIOwner, nil)
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf("%s*", *r.amiName),
		[]string{*r.amiOwner}, nil)

	if err != nil {
		return err
	}
	// Networking

	nw, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:                  *r.prefix,
			ID:                      awsWindowsDedicatedID,
			Region:                  *r.allocationData.Region,
			AZ:                      *r.allocationData.AZ,
			CreateLoadBalancer:      r.spot,
			Airgap:                  *r.airgap,
			AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
		})
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, awsWindowsDedicatedID, "pk")}
	keyResources, err := kpr.Create(ctx, r.mCtx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	// Security groups
	securityGroups, err := securityGroups(ctx, r.mCtx, r.prefix, nw.Vpc)
	if err != nil {
		return err
	}
	// Compute
	password, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, awsWindowsDedicatedID, "password"))
	if err != nil {
		return err
	}

	// Generate userdata
	userDataB64, err := cloudConfigWindowsServer.GenerateUserdata(ctx, &amiUserDefault, password, keyResources, r.mCtx.RunID())
	if err != nil {
		return err
	}

	cr := compute.ComputeRequest{
		MCtx:             r.mCtx,
		Prefix:           *r.prefix,
		ID:               awsWindowsDedicatedID,
		VPC:              nw.Vpc,
		Subnet:           nw.Subnet,
		AMI:              ami,
		UserDataAsBase64: userDataB64,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     requiredInstanceTypes,
		DiskSize:         &diskSize,
		Airgap:           *r.airgap,
		LB:               nw.LoadBalancer,
		Eip:              nw.Eip,
		LBTargetGroups:   []int{22, 3389}}
	if r.allocationData.SpotPrice != nil {
		cr.Spot = true
		cr.SpotPrice = *r.allocationData.SpotPrice
	}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername),
		pulumi.String(*r.amiUser))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPassword),
		password.Result)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost),
		c.GetHostDnsName(!*r.airgap))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx, r.mCtx,
			*r.allocationData.Region, *r.prefix,
			awsWindowsDedicatedID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"windows",
				r.mCtx.ProjectName(),
				r.mCtx.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	return c.Readiness(ctx, command.CommandPing, *r.prefix, awsWindowsDedicatedID,
		keyResources.PrivateKey, *r.amiUser, nw.Bastion, c.Dependencies)
}

// Write exported values in context to files o a selected target folder
func manageResults(mCtx *mc.Context, stackResult auto.UpResult, prefix *string, airgap *bool) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, outputUserPassword):   "userpassword",
		fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *prefix, outputHost):           "host",
	}
	if *airgap {
		err := bastion.WriteOutputs(stackResult, *prefix, mCtx.GetResultsOutputPath())
		if err != nil {
			return err
		}
	}
	return output.Write(stackResult, mCtx.GetResultsOutputPath(), results)
}

// security group for mac machine with ingress rules for ssh and vnc
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix *string,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// ingress for ssh access from 0.0.0.0
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	rdpIngressRule := securityGroup.RDP_TCP
	rdpIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(*prefix, awsWindowsDedicatedID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsWindowsDedicatedID),
		IngressRules: []securityGroup.IngressRules{
			sshIngressRule, rdpIngressRule},
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

// Need to add custom listener for RDP or should we use 22 tunneling through the bastion?
// func addCustomListeners(){}
