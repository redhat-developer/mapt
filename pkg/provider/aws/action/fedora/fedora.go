package fedora

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
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
	fedoraCloudConfig "github.com/redhat-developer/mapt/pkg/provider/util/cloud-config/fedora"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type FedoraArgs struct {
	Prefix         string
	Version        string
	Arch           string
	ComputeRequest *cr.ComputeRequestArgs
	Spot           bool
	Airgap         bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
}

type fedoraRequest struct {
	prefix         *string
	version        *string
	arch           *string
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

// Create orchestrate 2 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, args *FedoraArgs) (err error) {
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := fedoraRequest{
		prefix:  &prefix,
		version: &args.Version,
		arch:    &args.Arch,
		spot:    &args.Spot,
		timeout: &args.Timeout,
		airgap:  &args.Airgap}
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
	// if not only host the mac machine will be created
	if !*r.airgap {
		return r.createMachine()
	}
	// Airgap scneario requires orchestration
	return r.createAirgapMachine()
}

// Will destroy resources related to machine
func Destroy(ctx *maptContext.ContextArgs) (err error) {
	logging.Debug("Run fedora destroy")
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}

	// Destroy fedora related resources
	if err := aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: stackName,
		}); err != nil {
		return err
	}

	// Destroy spot orchestrated stack
	if spot.Exist() {
		return spot.Destroy()
	}
	return nil
}

func (r *fedoraRequest) createMachine() error {
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(stackName),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   maptContext.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        *r.allocationData.Region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.allocationData.Region}),
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
func (r *fedoraRequest) createAirgapMachine() error {
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
func (r *fedoraRequest) deploy(ctx *pulumi.Context) error {
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf(amiRegex[*r.arch], *r.version),
		[]string{amiOwner},
		map[string]string{
			"architecture": *r.arch})
	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix: *r.prefix,
		ID:     awsFedoraDedicatedID,
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
			*r.prefix, awsFedoraDedicatedID, "pk")}
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
	userDataB64, err := fedoraCloudConfig.Userdata(amiUserDefault)
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		Prefix:           *r.prefix,
		ID:               awsFedoraDedicatedID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		KeyResources:     keyResources,
		UserDataAsBase64: userDataB64,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.InstanceTypes,
		DiskSize:         &diskSize,
		Airgap:           *r.airgap,
		LB:               lb,
		LBEIP:            lbEIP,
		LBTargetGroups:   []int{22},
	}
	if *r.spot {
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
		c.GetHostIP(!*r.airgap))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx,
			*r.allocationData.Region, *r.prefix,
			awsFedoraDedicatedID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"fedora",
				maptContext.ProjectName(),
				maptContext.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	var cDeps []pulumi.Resource
	if *r.spot {
		cDeps = []pulumi.Resource{c.AutoscalingGroup}
	} else {
		cDeps = []pulumi.Resource{c.Instance}
	}
	return c.Readiness(ctx, command.CommandPing, *r.prefix, awsFedoraDedicatedID,
		keyResources.PrivateKey, amiUserDefault, bastion, cDeps)
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
	var ingressRules []securityGroup.IngressRules
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	ingressRules = []securityGroup.IngressRules{sshIngressRule}
	// Integration ports
	cirrusPort, err := cirrus.CirrusPort()
	if err != nil {
		return nil, err
	}
	if cirrusPort != nil {
		ingressRules = append(ingressRules,
			securityGroup.IngressRules{
				Description: fmt.Sprintf("Cirrus port for %s", awsFedoraDedicatedID),
				FromPort:    *cirrusPort,
				ToPort:      *cirrusPort,
				Protocol:    "tcp",
				CidrBlocks:  infra.NETWORKING_CIDR_ANY_IPV4,
			})
	}

	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:         resourcesUtil.GetResourceName(*prefix, awsFedoraDedicatedID, "sg"),
		VPC:          vpc,
		Description:  fmt.Sprintf("sg for %s", awsFedoraDedicatedID),
		IngressRules: ingressRules,
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
