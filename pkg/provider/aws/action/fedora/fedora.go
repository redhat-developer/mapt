package fedora

import (
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	infra "github.com/adrianriobo/qenvs/pkg/provider"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/bastion"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/ec2/compute"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/spot"
	amiSVC "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/ami"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/adrianriobo/qenvs/pkg/provider/aws/services/ec2/security-group"
	"github.com/adrianriobo/qenvs/pkg/provider/util/output"
	"github.com/adrianriobo/qenvs/pkg/util"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Request struct {
	Prefix  string
	Version string
	Spot    bool
	Airgap  bool
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

// Create orchestrate 2 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(r *Request) error {
	if r.Spot {
		sr := spot.SpotOptionRequest{
			Prefix:             r.Prefix,
			ProductDescription: "Linux/UNIX",
			InstaceTypes:       requiredInstanceTypes,
			AMIName:            fmt.Sprintf(amiRegex, r.Version),
			AMIArch:            "x86_64",
		}
		so, err := sr.Create()
		if err != nil {
			return err
		}
		r.region = so.Region
		r.az = so.AvailabilityZone
		r.spotPrice = so.MaxPrice
	} else {
		r.region = os.Getenv("AWS_DEFAULT_REGION")
		az, err := data.GetRandomAvailabilityZone(r.region)
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
func Destroy() (err error) {
	err = aws.DestroyStack(stackName)
	if err != nil {
		return
	}
	if spot.Exist() {
		return spot.Destroy()
	}
	return nil
}

func (r *Request) createMachine() error {
	cs := manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName(stackName),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: r.region}),
		DeployFunc: r.deploy,
	}

	csResult, err := manager.UpStack(cs)
	if err != nil {
		return err
	}
	err = r.manageResults(csResult)
	if err != nil {
		return err
	}
	return nil
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
		fmt.Sprintf(amiRegex, r.Version),
		amiOwner,
		map[string]string{
			"architecture": "x86_64"})
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
	cr := compute.ComputeRequest{
		Prefix:         r.Prefix,
		ID:             awsFedoraDedicatedID,
		VPC:            vpc,
		Subnet:         targetSubnet,
		AMI:            ami,
		KeyResources:   keyResources,
		SecurityGroups: securityGroups,
		InstaceTypes:   requiredInstanceTypes,
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
	return c.Readiness(ctx, r.Prefix, awsFedoraDedicatedID,
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
		err := bastion.WriteOutputs(stackResult, r.Prefix, qenvsContext.GetResultsOutputPath())
		if err != nil {
			return err
		}
	}
	return output.Write(stackResult, qenvsContext.GetResultsOutputPath(), results)
}

// security group for mac machine with ingress rules for ssh and vnc
func (r *Request) securityGroups(ctx *pulumi.Context,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// ingress for ssh access from 0.0.0.0
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(r.Prefix, awsFedoraDedicatedID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsFedoraDedicatedID),
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
