package fedora

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type Request struct {
	Prefix          string
	Version         string
	Arch            string
	InstanceRequest instancetypes.InstanceRequest
	VMType          []string
	Spot            bool
	Airgap          bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
	// internal management
	// For airgap scenario there is an orchestation of
	// a phase with connectivity on the machine (allowing bootstraping)
	// a pahase with connectivyt off where the subnet for the target lost the nat gateway
	airgapPhaseConnectivity network.Connectivity
	// location and price (if Spot is enable)
	region    string
	az        string
	spotPrice float64
}

type userDataValues struct {
	Username             string
	ActionsRunnerSnippet string
	CirrusSnippet        string
}

//go:embed cloud-config-base
var CloudConfigBase []byte

// Create orchestrate 2 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, r *Request) error {
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
		return err
	}

	if len(r.VMType) == 0 {
		vmTypes, err := r.InstanceRequest.GetMachineTypes()
		if err != nil {
			logging.Debugf("Unable to fetch desire instance types: %v", err)
		}
		if len(vmTypes) > 0 {
			r.VMType = append(r.VMType, vmTypes...)
		}
	}
	if r.Spot {
		sr := spot.SpotOptionRequest{
			Prefix:             r.Prefix,
			ProductDescription: "Linux/UNIX",
			InstaceTypes: util.If(len(r.VMType) > 0,
				r.VMType,
				supportedInstanceTypes[r.Arch]),
			AMIName: fmt.Sprintf(amiRegex[r.Arch], r.Version),
			AMIArch: r.Arch,
		}
		so, err := sr.Create()
		if err != nil {
			return err
		}
		r.region = so.Region
		r.az = so.AvailabilityZone
		r.spotPrice = so.MaxPrice
		r.VMType, err = data.FilterInstaceTypesOfferedByRegion(r.VMType, r.region)
		if err != nil {
			return err
		}
	} else {
		r.region = os.Getenv("AWS_DEFAULT_REGION")
		az, err := data.GetRandomAvailabilityZone(r.region, nil)
		if err != nil {
			return err
		}
		r.az = *az
	}

	// if not only host the mac machine will be created
	if !r.Airgap {
		return r.createMachine()
	}
	// Airgap scneario requires orchestration
	return r.createAirgapMachine()
}

// Will destroy resources related to machine
func Destroy(ctx *maptContext.ContextArgs) (err error) {
	logging.Debug("Run fedora destroy")
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
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

func (r *Request) createMachine() error {
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(stackName),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   maptContext.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        r.region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: r.region}),
		DeployFunc: r.deploy,
	}

	sr, _ := manager.UpStack(cs)
	return r.manageResults(sr)
}

// Abstract this with a stackAirgapHandle receives a fn (connectivty on / off) err executes
// first on then off
func (r *Request) createAirgapMachine() error {
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
func (r *Request) deploy(ctx *pulumi.Context) error {
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf(amiRegex[r.Arch], r.Version),
		amiOwner,
		map[string]string{
			"architecture": r.Arch})
	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix: r.Prefix,
		ID:     awsFedoraDedicatedID,
		Region: r.region,
		AZ:     r.az,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer:      &r.Spot,
		Airgap:                  r.Airgap,
		AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
	}
	vpc, targetSubnet, _, bastion, lb, err := nr.Network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			r.Prefix, awsFedoraDedicatedID, "pk")}
	keyResources, err := kpr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	// Security groups
	securityGroups, err := r.securityGroups(ctx, vpc)
	if err != nil {
		return err
	}
	userDataB64, err := r.getUserdata()
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		Prefix:           r.Prefix,
		ID:               awsFedoraDedicatedID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		KeyResources:     keyResources,
		UserDataAsBase64: userDataB64,
		SecurityGroups:   securityGroups,
		InstaceTypes: util.If(len(r.VMType) > 0,
			r.VMType,
			supportedInstanceTypes[r.Arch]),
		DiskSize:       &diskSize,
		Airgap:         r.Airgap,
		LB:             lb,
		LBTargetGroups: []int{22},
		Spot:           r.Spot}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUsername),
		pulumi.String(amiUserDefault))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost),
		c.GetHostIP(!r.Airgap))
	if len(r.Timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx,
			r.region, r.Prefix,
			awsFedoraDedicatedID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"fedora",
				maptContext.ProjectName(),
				maptContext.BackedURL()),
			r.Timeout); err != nil {
			return err
		}
	}
	return c.Readiness(ctx, command.CommandPing, r.Prefix, awsFedoraDedicatedID,
		keyResources.PrivateKey, amiUserDefault, bastion, []pulumi.Resource{})
}

// Write exported values in context to files o a selected target folder
func (r *Request) manageResults(stackResult auto.UpResult) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):           "host",
	}
	if r.Airgap {
		err := bastion.WriteOutputs(stackResult, r.Prefix, maptContext.GetResultsOutputPath())
		if err != nil {
			return err
		}
	}
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), results)
}

// security group for mac machine with ingress rules for ssh and vnc
func (r *Request) securityGroups(ctx *pulumi.Context,
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
		Name:         resourcesUtil.GetResourceName(r.Prefix, awsFedoraDedicatedID, "sg"),
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

func (r *Request) getUserdata() (pulumi.StringPtrInput, error) {
	cirrusSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(cirrus.GetRunnerArgs(), amiUserDefault)
	if err != nil {
		return nil, err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippetAsCloudInitWritableFile(github.GetRunnerArgs(), amiUserDefault)
	if err != nil {
		return nil, err
	}

	templateConfig := string(CloudConfigBase[:])
	userdata, err := file.Template(
		userDataValues{
			amiUserDefault,
			*ghActionsRunnerSnippet,
			*cirrusSnippet},
		templateConfig)
	return pulumi.String(base64.StdEncoding.EncodeToString([]byte(userdata))), err
}
