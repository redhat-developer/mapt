package openshiftsnc

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	amiCopy "github.com/redhat-developer/mapt/pkg/provider/aws/modules/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ssm"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type OpenshiftSNCArgs struct {
	Prefix          string
	InstanceRequest instancetypes.InstanceRequest
	Version         string
	Arch            string
	PullSecretFile  string
	CaCertFile      string
	Spot            bool
	Timeout         string
}

type allocationData struct {
	// location and price (if Spot is enable)
	region        *string
	az            *string
	spotPrice     *float64
	instanceTypes []string
}

type openshiftSNCRequest struct {
	prefix         *string
	version        *string
	arch           *string
	timeout        *string
	pullSecretFile *string
	caCertFile     *string
	allocationData *allocationData
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, args *OpenshiftSNCArgs) error {
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
	r := openshiftSNCRequest{
		prefix:         &prefix,
		version:        &args.Version,
		arch:           &args.Arch,
		pullSecretFile: &args.PullSecretFile,
		caCertFile:     &args.CaCertFile,
		timeout:        &args.Timeout}
	r.allocationData, err = util.IfWithError(args.Spot,
		func() (*allocationData, error) {
			return getSpotAllocationData(&args.Prefix,
				&args.Version, &args.Arch, instanceTypes)
		},
		func() (*allocationData, error) {
			return getDefaultAllocationData()
		})
	if err != nil {
		return err
	}
	// Manage AMI offering / replication
	amiName := amiName(&args.Version, &args.Arch)
	if err = manageAMIReplication(&args.Prefix,
		&amiName, r.allocationData.region, &args.Arch); err != nil {
		return err
	}
	return r.createCluster()
}

// Will destroy resources related to machine
func Destroy(ctx *maptContext.ContextArgs) (err error) {
	logging.Debug("Run openshift destroy")
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

func (r *openshiftSNCRequest) createCluster() error {
	cs := manager.Stack{
		StackName:   maptContext.StackNameByProject(stackName),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   maptContext.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        *r.allocationData.region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.allocationData.region}),
		DeployFunc: r.deploy,
	}
	sr, _ := manager.UpStack(cs)
	return manageResults(sr, r.prefix)
}

func (r *openshiftSNCRequest) deploy(ctx *pulumi.Context) error {
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		amiName(r.version, r.arch),
		[]string{"self", amiOwner},
		map[string]string{
			"architecture": *r.arch})
	if err != nil {
		return err
	}
	// Networking
	lbEnable := true
	nr := network.NetworkRequest{
		Prefix: *r.prefix,
		ID:     awsOCPSNCID,
		Region: *r.allocationData.region,
		AZ:     *r.allocationData.az,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer: &lbEnable,
		Airgap:             false,
	}
	vpc, targetSubnet, _, bastion, lb, err := nr.Network(ctx)
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, awsOCPSNCID, "pk")}
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
	// Userdata
	udB64, udDependecies, err := r.userData(ctx)
	if err != nil {
		return err
	}
	// Create instance
	cr := compute.ComputeRequest{
		Prefix:           *r.prefix,
		ID:               awsOCPSNCID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.instanceTypes,
		DiskSize:         &diskSize,
		LB:               lb,
		LBTargetGroups:   []int{securityGroup.SSH_PORT, portHTTPS, portAPI},
		SpotPrice:        strconv.FormatFloat(*r.allocationData.spotPrice, 'f', -1, 64),
		Spot:             true,
		UserDataAsBase64: udB64,
		DependsOn:        udDependecies,
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
		if err = serverless.OneTimeDelayedTask(ctx,
			*r.allocationData.region, *r.prefix,
			awsOCPSNCID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"openshift-snc",
				maptContext.ProjectName(),
				maptContext.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	//TODO need to get the kubeconfig
	kubeconfig, err := kubeconfig(ctx, r.prefix, c, keyResources.PrivateKey)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig),
		kubeconfig)
	return c.Readiness(ctx, command.CommandPing, *r.prefix, awsOCPSNCID,
		keyResources.PrivateKey, amiUserDefault, bastion, []pulumi.Resource{})
}

// Write exported values in context to files o a selected target folder
func manageResults(stackResult auto.UpResult, prefix *string) error {
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *prefix, outputHost):           "host",
		fmt.Sprintf("%s-%s", *prefix, outputKubeconfig):     "kubeconfig",
	}
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), results)
}

// security group for Openshift
func securityGroups(ctx *pulumi.Context, prefix *string,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(*prefix, awsOCPSNCID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsOCPSNCID),
		IngressRules: []securityGroup.IngressRules{securityGroup.SSH_TCP,
			{Description: "Console", FromPort: portHTTPS, ToPort: portHTTPS, Protocol: "tcp"},
			{Description: "API", FromPort: portAPI, ToPort: portAPI, Protocol: "tcp"}},
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

func getSpotAllocationData(prefix *string,
	version, arch *string, instanceTypes []string) (*allocationData, error) {
	sr := spot.SpotOptionRequest{
		Prefix:             *prefix,
		ProductDescription: amiProductDescription,
		InstaceTypes:       instanceTypes,
		AMIName:            amiName(version, arch),
		AMIArch:            *arch,
	}
	so, err := sr.Create()
	if err != nil {
		return nil, err
	}
	availableInstaceTypes, err :=
		data.FilterInstaceTypesOfferedByRegion(instanceTypes, so.Region)
	if err != nil {
		return nil, err
	}
	return &allocationData{
		region:        &so.Region,
		az:            &so.AvailabilityZone,
		spotPrice:     &so.MaxPrice,
		instanceTypes: availableInstaceTypes,
	}, nil
}

func getDefaultAllocationData() (ad *allocationData, err error) {
	ad = &allocationData{}
	region := os.Getenv("AWS_DEFAULT_REGION")
	ad.region = &region
	ad.az, err = data.GetRandomAvailabilityZone(region, nil)
	return
}

func manageAMIReplication(prefix, amiName, region, arch *string) error {
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
			Prefix:          *prefix,
			ID:              awsOCPSNCID,
			AMISourceName:   amiName,
			AMISourceArch:   arch,
			AMITargetRegion: &amiOriginRegion,
			// TODO add this as param
			AMIKeepCopy: false,
		}
		if err := acr.Create(); err != nil {
			return err
		}
	}
	return nil
}

func (r *openshiftSNCRequest) userData(ctx *pulumi.Context) (pulumi.StringPtrInput, []pulumi.Resource, error) {
	// Resources
	dependecies := []pulumi.Resource{}
	// Manage pull secret
	ps, err := os.ReadFile(*r.pullSecretFile)
	if err != nil {
		return nil, nil, err
	}
	psB64 := base64.StdEncoding.EncodeToString([]byte(ps))
	psName, psParam, err := ssm.AddSSM(ctx, r.prefix, &ocpPullSecretID, &psB64)
	if err != nil {
		return nil, nil, err
	}
	dependecies = append(dependecies, psParam)
	// Manage ca crt
	ca, err := os.ReadFile(*r.caCertFile)
	if err != nil {
		return nil, nil, err
	}
	caB64 := base64.StdEncoding.EncodeToString([]byte(ca))
	caName, caParam, err := ssm.AddSSM(ctx, r.prefix, &cacertID, &caB64)
	if err != nil {
		return nil, nil, err
	}
	dependecies = append(dependecies, caParam)
	// KubeAdmin pass
	kaPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, awsOCPSNCID, "kubeadminpassword"))
	if err != nil {
		return nil, nil, err
	}
	kaPassName, kaPassParam, err := ssm.AddSSMFromResource(ctx, r.prefix, &kapass, kaPassword.Result)
	if err != nil {
		return nil, nil, err
	}
	dependecies = append(dependecies, kaPassParam)
	// Developer pass
	devPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, awsOCPSNCID, "devpassword"))
	if err != nil {
		return nil, nil, err
	}
	devPassName, devPassParam, err := ssm.AddSSMFromResource(ctx, r.prefix, &devpass, devPassword.Result)
	if err != nil {
		return nil, nil, err
	}
	dependecies = append(dependecies, devPassParam)
	ccB64, err := cloudConfig(dataValues{
		SSMPullSecretName:        *psName,
		SSMCaCertName:            *caName,
		SSMKubeAdminPasswordName: *kaPassName,
		SSMDeveloperPasswordName: *devPassName})
	return pulumi.String(*ccB64), dependecies, err
}

func kubeconfig(ctx *pulumi.Context, prefix *string, c *compute.Compute, mk *tls.PrivateKey) (pulumi.StringOutput, error) {
	// Once the cluster setup is comleted we
	// get the kubeconfig file from the host running the cluster
	// then we replace the internal access with the public IP
	// the resulting kubeconfig file can be used to access the cluster
	getKCCmd := ("cat /opt/kubeconfig")
	getKC, err := c.RunCommand(ctx, getKCCmd, *prefix, awsOCPSNCID, mk, amiUserDefault, nil, nil)
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	kubeconfig := pulumi.All(getKC.Stdout, c.Instance.PublicIp).ApplyT(
		func(args []interface{}) string {
			return strings.ReplaceAll(args[0].(string),
				"https://api.crc.testing:6443",
				fmt.Sprintf("https://api.%s.nip.io:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kubeconfig, nil
}
