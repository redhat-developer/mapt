package rhel

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	cloudConfigRHEL "github.com/redhat-developer/mapt/pkg/provider/util/cloud-config/rhel"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type RHELArgs struct {
	Prefix          string
	Version         string
	Arch            string
	InstanceRequest instancetypes.InstanceRequest
	SubsUsername    string
	SubsUserpass    string
	ProfileSNC      bool
	Spot            bool
	Airgap          bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
}

type rhelRequest struct {
	prefix         *string
	version        *string
	arch           *string
	spot           *bool
	subsUsername   *string
	subsUserpass   *string
	profileSNC     *bool
	timeout        *string
	allocationData *allocation.AllocationData
	airgap         *bool
	// internal management
	// For airgap scenario there is an orchestation of
	// a phase with connectivity on the machine (allowing bootstraping)
	// a pahase with connectivyt off where the subnet for the target lost the nat gateway
	airgapPhaseConnectivity network.Connectivity
}

// Create orchestrate 2 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, args *RHELArgs) error {
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Get instance types matching requirements
	instanceTypes, err := args.InstanceRequest.GetMachineTypes()
	if err != nil {
		return err
	}
	if len(instanceTypes) == 0 {
		return fmt.Errorf("no instances matching criteria")
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := rhelRequest{
		prefix:       &prefix,
		version:      &args.Version,
		arch:         &args.Arch,
		spot:         &args.Spot,
		timeout:      &args.Timeout,
		subsUsername: &args.SubsUsername,
		subsUserpass: &args.SubsUserpass,
		profileSNC:   &args.ProfileSNC,
		airgap:       &args.Airgap}
	r.allocationData, err = util.IfWithError(args.Spot,
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnSpot(
				&args.Prefix, &amiProduct, nil, instanceTypes)
		},
		func() (*allocation.AllocationData, error) {
			return allocation.AllocationDataOnDemand()
		})
	if err != nil {
		return err
	}
	// if not only host the mac machine will be created
	if !*r.airgap {
		return r.createMachine()
	}
	// Airgap scneario requires orchestration
	return r.createAirgapMachine()
}

// Will destroy resources related to machine
func Destroy(ctx *maptContext.ContextArgs) error {
	logging.Debug("Run rhel destroy")
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
	if spot.Exist() {
		return spot.Destroy()
	}
	return nil
}

func (r *rhelRequest) createMachine() error {
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
func (r *rhelRequest) createAirgapMachine() error {
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
func (r *rhelRequest) deploy(ctx *pulumi.Context) error {
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf(amiRegex, *r.version, *r.arch),
		nil,
		map[string]string{
			"architecture": *r.arch})
	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix: *r.prefix,
		ID:     awsRHELDedicatedID,
		Region: *r.allocationData.Region,
		AZ:     *r.allocationData.AZ,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer:      r.spot,
		Airgap:                  *r.airgap,
		AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
	}
	vpc, targetSubnet, _, bastion, lb, lbEIP, err := nr.Network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, awsRHELDedicatedID, "pk")}
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
	rhelCloudConfig := &cloudConfigRHEL.RequestArgs{
		SNCProfile:   *r.profileSNC,
		SubsUsername: *r.subsUsername,
		SubsPassword: *r.subsUserpass,
		Username:     amiUserDefault}
	userDataB64, err := rhelCloudConfig.GetAsUserdata()
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		Prefix:           *r.prefix,
		ID:               awsRHELDedicatedID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		UserDataAsBase64: pulumi.String(userDataB64),
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.InstanceTypes,
		DiskSize:         &diskSize,
		Airgap:           *r.airgap,
		LB:               lb,
		LBEIP:            lbEIP,
		LBTargetGroups:   []int{22},
		Spot:             *r.spot}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername),
		pulumi.String(amiUserDefault))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost),
		c.GetHostIP(!*r.airgap))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(
			ctx,
			fmt.Sprintf("rhel-timeout-%s", maptContext.RunID()),
			*r.allocationData.Region, *r.prefix,
			awsRHELDedicatedID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"rhel",
				maptContext.ProjectName(),
				maptContext.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	return c.Readiness(ctx, command.CommandCloudInitWait, *r.prefix, awsRHELDedicatedID,
		keyResources.PrivateKey, amiUserDefault, bastion, []pulumi.Resource{})
}

// Write exported values in context to files o a selected target folder
func manageResults(stackResult auto.UpResult, prefix *string, airgap *bool) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
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
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(*prefix, awsRHELDedicatedID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsRHELDedicatedID),
		IngressRules: []securityGroup.IngressRules{
			sshIngressRule},
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
