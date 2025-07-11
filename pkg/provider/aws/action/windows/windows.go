package windows

import (
	_ "embed"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
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
	Spot           bool
	Airgap         bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
}

type windowsServerRequest struct {
	prefix *string

	amiName     *string
	amiUser     *string
	amiOwner    *string
	amiLang     *string
	amiKeepCopy *bool

	spot           *bool
	timeout        *string
	allocationData *allocation.AllocationData
	airgap         *bool
	// internal management
	// For airgap scenario there is an orchestation of
	// a phase with connectivity on the machine (allowing bootstraping)
	// a pahase with connectivyt off where the subnet for the target lost the nat gateway
	airgapPhaseConnectivity network.Connectivity
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, args *WindowsServerArgs) (err error) {
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
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
		prefix:      &prefix,
		amiName:     &args.AMIName,
		amiUser:     &args.AMIUser,
		amiOwner:    &args.AMIOwner,
		amiKeepCopy: &args.AMIKeepCopy,
		amiLang:     &args.AMILang,
		spot:        &args.Spot,
		timeout:     &args.Timeout,
		airgap:      &args.Airgap}
	r.allocationData, err = util.IfWithError(args.Spot,
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnSpot(
				&args.Prefix, &amiProduct, nil, args.ComputeRequest)
		},
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnDemand()
		})
	if err != nil {
		return err
	}

	isAMIOffered, _, err := data.IsAMIOffered(
		data.ImageRequest{
			Name:   r.amiName,
			Region: r.allocationData.Region})
	if err != nil {
		return err
	}
	// If it is not offered need to create a copy on the target region
	if !isAMIOffered {
		acr := amiCopy.CopyAMIRequest{
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
func Destroy(ctx *maptContext.ContextArgs) (err error) {
	logging.Debug("Run windows destroy")
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}

	if err := aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: stackName,
		}); err != nil {
		return err
	}
	if amiCopy.Exist() {
		err = amiCopy.Destroy()
		if err != nil {
			return
		}
	}
	if spot.Exist() {
		return spot.Destroy()
	}
	return nil
}

func (r *windowsServerRequest) createMachine() error {
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(stackName),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   maptContext.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION: *r.allocationData.Region}),
		DeployFunc: r.deploy,
	}

	sr, err := manager.UpStack(cs)
	if err != nil {
		return err
	}
	return manageResults(sr, r.prefix, r.airgap)
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
	// Get AMI ref
	// ami, err := amiSVC.GetAMIByName(ctx, r.AMIName, r.AMIOwner, nil)
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf("%s*", *r.amiName),
		[]string{*r.amiOwner}, nil)

	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix: *r.prefix,
		ID:     awsWindowsDedicatedID,
		Region: *r.allocationData.Region,
		AZ:     *r.allocationData.AZ,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer:      r.spot,
		Airgap:                  *r.airgap,
		AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
	}
	// vpc, targetSubnet, targetRouteTableAssociation, bastion, lb, err := nr.Network(ctx)
	vpc, targetSubnet, _, bastion, lb, lbEIP, err := nr.Network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, awsWindowsDedicatedID, "pk")}
	keyResources, err := kpr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	// Security groups
	securityGroups, err := securityGroups(ctx, r.prefix, vpc)
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
	userDataB64, err := cloudConfigWindowsServer.Userdata(ctx, &amiUserDefault, password, keyResources)
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		Prefix:           *r.prefix,
		ID:               awsWindowsDedicatedID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		UserDataAsBase64: userDataB64,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     requiredInstanceTypes,
		DiskSize:         &diskSize,
		Airgap:           *r.airgap,
		LB:               lb,
		LBEIP:            lbEIP,
		LBTargetGroups:   []int{22, 3389},
		Spot:             *r.spot}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername),
		pulumi.String(*r.amiUser))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPassword),
		password.Result)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost),
		c.GetHostIP(!*r.airgap))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx,
			*r.allocationData.Region, *r.prefix,
			awsWindowsDedicatedID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"windows",
				maptContext.ProjectName(),
				maptContext.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	return c.Readiness(ctx, command.CommandPing, *r.prefix, awsWindowsDedicatedID,
		keyResources.PrivateKey, *r.amiUser, bastion, c.Dependencies)
}

// Write exported values in context to files o a selected target folder
func manageResults(stackResult auto.UpResult, prefix *string, airgap *bool) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, outputUserPassword):   "userpassword",
		fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *prefix, outputHost):           "host",
	}
	if *airgap {
		err := bastion.WriteOutputs(stackResult, *prefix, maptContext.GetResultsOutputPath())
		if err != nil {
			return err
		}
	}
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), results)
}

// security group for mac machine with ingress rules for ssh and vnc
func securityGroups(ctx *pulumi.Context, prefix *string,
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
	}.Create(ctx)
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
