package windows

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	amiCopy "github.com/redhat-developer/mapt/pkg/provider/aws/modules/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/bastion"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

// add proxy https://github.com/ptcodes/proxy-server-with-terraform/blob/master/main.tf
type Request struct {
	Prefix string
	// AMI info. Optional. User and Owner only applied
	// if AMIName is set
	AMIName     string
	AMIUser     string
	AMIOwner    string
	AMILang     string
	AMIKeepCopy bool
	// Features
	Spot   bool
	Airgap bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
	// setup as github actions runner
	SetupGHActionsRunner bool
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
	Password             string
	AuthorizedKey        string
	InstallActionsRunner bool
	ActionsRunnerSnippet string
	RunnerToken          string
}

//go:embed bootstrap.ps1
var BootstrapScript []byte

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, r *Request) error {
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
		return err
	}

	if len(r.AMIName) == 0 {
		r.AMIName = amiNameDefault
		r.AMIUser = amiUserDefault
		r.AMIOwner = amiOwnerDefault
	}
	if len(r.AMILang) > 0 && r.AMILang == amiLangNonEng {
		r.AMIName = amiNonEngNameDefault
	}
	if r.Spot {
		// On windows we use a custom AMI as so we
		// do not add it as a requirement for best spot option, we will get the region
		// and then wil replicate AMI if needed
		sr := spot.SpotOptionRequest{
			Prefix:             r.Prefix,
			ProductDescription: "Windows",
			InstaceTypes:       requiredInstanceTypes,
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
		az, err := data.GetRandomAvailabilityZone(r.region, nil)
		if err != nil {
			return err
		}
		r.az = *az
	}
	isAMIOffered, _, err := data.IsAMIOffered(
		data.ImageRequest{
			Name:   &r.AMIName,
			Region: &r.region})
	if err != nil {
		return err
	}
	// If it is not offered need to create a copy on the target region
	if !isAMIOffered {
		acr := amiCopy.CopyAMIRequest{
			Prefix:          r.Prefix,
			ID:              awsWindowsDedicatedID,
			AMISourceName:   &r.AMIName,
			AMISourceArch:   nil,
			AMITargetRegion: &r.region,
			AMIKeepCopy:     r.AMIKeepCopy,
			FastLaunch:      amiFastLaunch,
			MaxParallel:     int32(amiFastLaunchMaxParallel),
		}
		if err := acr.Create(); err != nil {
			return err
		}
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
	logging.Debug("Run windows destroy")
	// Create mapt Context
	if err := maptContext.Init(ctx); err != nil {
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

func (r *Request) createMachine() error {
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(stackName),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   maptContext.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: r.region}),
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
	// Get AMI ref
	// ami, err := amiSVC.GetAMIByName(ctx, r.AMIName, r.AMIOwner, nil)
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf("%s*", r.AMIName),
		r.AMIOwner, nil)

	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix: r.Prefix,
		ID:     awsWindowsDedicatedID,
		Region: r.region,
		AZ:     r.az,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer:      &r.Spot,
		Airgap:                  r.Airgap,
		AirgapPhaseConnectivity: r.airgapPhaseConnectivity,
	}
	// vpc, targetSubnet, targetRouteTableAssociation, bastion, lb, err := nr.Network(ctx)
	vpc, targetSubnet, _, bastion, lb, err := nr.Network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			r.Prefix, awsWindowsDedicatedID, "pk")}
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
	// Compute
	password, userDataB64, err := r.getUserdata(ctx, keyResources)
	if err != nil {
		return err
	}
	cr := compute.ComputeRequest{
		Prefix:           r.Prefix,
		ID:               awsWindowsDedicatedID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		UserDataAsBase64: userDataB64,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     requiredInstanceTypes,
		DiskSize:         &diskSize,
		Airgap:           r.Airgap,
		LB:               lb,
		LBTargetGroups:   []int{22, 3389},
		Spot:             r.Spot}
	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUsername),
		pulumi.String(r.AMIUser))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword),
		password.Result)
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost),
		c.GetHostIP(!r.Airgap))
	if len(r.Timeout) > 0 {
		if err = serverless.CreateDestroyOperation(ctx,
			r.region, r.Prefix, awsWindowsDedicatedID,
			"windows", r.Timeout); err != nil {
			return err
		}
	}
	return c.Readiness(ctx, command.CommandPing, r.Prefix, awsWindowsDedicatedID,
		keyResources.PrivateKey, r.AMIUser, bastion, []pulumi.Resource{})
}

// Write exported values in context to files o a selected target folder
func (r *Request) manageResults(stackResult auto.UpResult) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword):   "userpassword",
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
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	rdpIngressRule := securityGroup.RDP_TCP
	rdpIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(r.Prefix, awsWindowsDedicatedID, "sg"),
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

// function to template userdata script to be executed on boot
func (r *Request) getUserdata(ctx *pulumi.Context,
	keypair *keypair.KeyPairResources) (
	*random.RandomPassword, pulumi.StringPtrInput, error) {
	password, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			r.Prefix, awsWindowsDedicatedID, "password"))
	if err != nil {
		return nil, nil, err
	}
	udBase64 := pulumi.All(password.Result, keypair.PrivateKey.PublicKeyOpenssh).ApplyT(
		func(args []interface{}) (string, error) {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			udv := userDataValues{
				r.AMIUser,
				password,
				authorizedKey,
				r.SetupGHActionsRunner,
				github.GetActionRunnerSnippetWin(),
				github.GetToken(),
			}
			userdata, err := file.Template(udv, string(BootstrapScript[:]))
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
		}).(pulumi.StringOutput)
	return password, udBase64, nil
}

// Need to add custom listener for RDP or should we use 22 tunneling through the bastion?
// func addCustomListeners(){}
