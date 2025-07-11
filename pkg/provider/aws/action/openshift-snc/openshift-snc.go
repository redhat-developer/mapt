package openshiftsnc

import (
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	amiCopy "github.com/redhat-developer/mapt/pkg/provider/aws/modules/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/iam"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ssm"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type OpenshiftSNCArgs struct {
	Prefix         string
	ComputeRequest *cr.ComputeRequestArgs
	Version        string
	Arch           string
	PullSecretFile string
	Spot           bool
	Timeout        string
}

type openshiftSNCRequest struct {
	prefix         *string
	version        *string
	arch           *string
	timeout        *string
	pullSecretFile *string
	allocationData *allocation.AllocationData
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(ctx *maptContext.ContextArgs, args *OpenshiftSNCArgs) (err error) {
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := openshiftSNCRequest{
		prefix:         &prefix,
		version:        &args.Version,
		arch:           &args.Arch,
		pullSecretFile: &args.PullSecretFile,
		timeout:        &args.Timeout}
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
	// Manage AMI offering / replication
	amiName := amiName(&args.Version, &args.Arch)
	if err = manageAMIReplication(&args.Prefix,
		&amiName, r.allocationData.Region, &args.Arch); err != nil {
		return err
	}
	return r.createCluster()
}

// Will destroy resources related to machine
func Destroy(ctx *maptContext.ContextArgs) (err error) {
	logging.Debug("Run openshift destroy")
	// Create mapt Context
	if err = maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	// Destroy fedora related resources
	if err = aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: stackName,
		}); err != nil {
		return err
	}
	// AMI Copy
	if amiCopy.Exist() {
		err = amiCopy.Destroy()
		if err != nil {
			return
		}
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
				awsConstants.CONFIG_AWS_REGION:        *r.allocationData.Region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.allocationData.Region}),
		DeployFunc: r.deploy,
	}
	sr, _ := manager.UpStack(cs)
	return manageResults(sr, r.prefix)
}

func (r *openshiftSNCRequest) deploy(ctx *pulumi.Context) error {
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf("%s*", amiName(r.version, r.arch)),
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
		Region: *r.allocationData.Region,
		AZ:     *r.allocationData.AZ,
		// LB is required if we use as which is used for spot feature
		CreateLoadBalancer: &lbEnable,
		Airgap:             false,
	}
	vpc, targetSubnet, _, _, lb, lbEIP, err := nr.Network(ctx)
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
	if maptContext.Debug() {
		keyResources.PrivateKey.PrivateKeyPem.ApplyT(
			func(privateKey string) (*string, error) {
				logging.Debugf("%s", privateKey)
				return nil, nil
			})
	}
	// Security groups
	securityGroups, err := securityGroups(ctx, r.prefix, vpc)
	if err != nil {
		return err
	}
	// Instance profile required by logic within userdata
	iProfile, err := iam.InstanceProfile(ctx, r.prefix, &awsOCPSNCID, cloudConfigRequiredProfiles)
	if err != nil {
		return err
	}
	// Userdata
	udB64, kaPass, devPass, udDependecies, err := r.userData(ctx, &keyResources.PrivateKey.PublicKeyOpenssh, &lbEIP.PublicIp)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeAdminPass),
		kaPass)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputDeveloperPass),
		devPass)
	// Create instance
	cr := compute.ComputeRequest{
		Prefix:           *r.prefix,
		ID:               awsOCPSNCID,
		VPC:              vpc,
		Subnet:           targetSubnet,
		AMI:              ami,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.InstanceTypes,
		DiskSize:         &diskSize,
		LB:               lb,
		LBEIP:            lbEIP,
		LBTargetGroups:   []int{securityGroup.SSH_PORT, portHTTPS, portAPI},
		SpotPrice:        *r.allocationData.SpotPrice,
		Spot:             true,
		InstanceProfile:  iProfile,
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
			*r.allocationData.Region, *r.prefix,
			awsOCPSNCID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"openshift-snc",
				maptContext.ProjectName(),
				maptContext.BackedURL()),
			*r.timeout); err != nil {
			return err
		}
	}
	// Use kubeconfig as the readiness for the cluster
	kubeconfig, err := kubeconfig(ctx, r.prefix, c, keyResources.PrivateKey)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig),
		pulumi.ToSecret(kubeconfig))
	return nil
}

// Write exported values in context to files o a selected target folder
func manageResults(stackResult auto.UpResult, prefix *string) error {
	hostIPKey := fmt.Sprintf("%s-%s", *prefix, outputHost)
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
		hostIPKey: "host",
		fmt.Sprintf("%s-%s", *prefix, outputKubeconfig):    "kubeconfig",
		fmt.Sprintf("%s-%s", *prefix, outputKubeAdminPass): "kubeadmin_pass",
		fmt.Sprintf("%s-%s", *prefix, outputDeveloperPass): "developer_pass",
	}
	if err := output.Write(
		stackResult, maptContext.GetResultsOutputPath(), results); err != nil {
		return err
	}
	eip, ok := stackResult.Outputs[hostIPKey].Value.(string)
	if ok {
		fmt.Printf("Cluster has been started you can access console at: %s. You can check passwords at %s",
			fmt.Sprintf(consoleURLRegex, eip),
			maptContext.GetResultsOutputPath())
		return nil
	}
	return fmt.Errorf("error getting value for cluster ip")
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

func (r *openshiftSNCRequest) userData(ctx *pulumi.Context,
	newPublicKey, lbEIP *pulumi.StringOutput,
) (pulumi.StringPtrInput, pulumi.StringInput, pulumi.StringInput, []pulumi.Resource, error) {
	// Resources
	dependecies := []pulumi.Resource{}
	// Manage pull secret
	ps, err := os.ReadFile(*r.pullSecretFile)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	psString := string(ps)
	psName, psParam, err := ssm.AddSSM(ctx, r.prefix, &ocpPullSecretID, &psString)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dependecies = append(dependecies, psParam)
	// KubeAdmin pass
	kaPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, awsOCPSNCID, "kubeadminpassword"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	kaPassName, kaPassParam, err := ssm.AddSSMFromResource(ctx, r.prefix, &kapass, kaPassword.Result)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dependecies = append(dependecies, kaPassParam)
	// Developer pass
	devPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, awsOCPSNCID, "devpassword"))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	devPassName, devPassParam, err := ssm.AddSSMFromResource(ctx, r.prefix, &devpass, devPassword.Result)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dependecies = append(dependecies, devPassParam)
	ccB64 := pulumi.All(newPublicKey, lbEIP).ApplyT(
		func(args []interface{}) (string, error) {
			ccB64, err := cloudConfig(dataValues{
				Username:                 amiUserDefault,
				PubKey:                   args[0].(string),
				PublicIP:                 args[1].(string),
				SSMPullSecretName:        *psName,
				SSMKubeAdminPasswordName: *kaPassName,
				SSMDeveloperPasswordName: *devPassName})
			return *ccB64, err
		}).(pulumi.StringOutput)

	return ccB64, kaPassword.Result, devPassword.Result, dependecies, err
}

func kubeconfig(ctx *pulumi.Context,
	prefix *string,
	c *compute.Compute, mk *tls.PrivateKey,
) (pulumi.StringOutput, error) {
	// Once the cluster setup is comleted we
	// get the kubeconfig file from the host running the cluster
	// then we replace the internal access with the public IP
	// the resulting kubeconfig file can be used to access the cluster

	// Check cluster is ready
	ocpReadyCmd, err := c.RunCommand(ctx,
		commandReadiness,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ocp-readiness", *prefix), awsOCPSNCID,
		mk, amiUserDefault, nil, nil)
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	// Check ocp-cluster-ca.service succeeds
	ocpCaRotatedCmd, err := c.RunCommand(ctx,
		commandCaServiceRan,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ocp-ca-rotated", *prefix), awsOCPSNCID,
		mk, amiUserDefault, nil, []pulumi.Resource{ocpReadyCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Get content for /opt/kubeconfig
	getKCCmd := ("cat /opt/kubeconfig")
	getKC, err := c.RunCommand(ctx,
		getKCCmd,
		compute.NoLoggingCmdStd,
		fmt.Sprintf("%s-kubeconfig", *prefix), awsOCPSNCID, mk, amiUserDefault,
		nil, []pulumi.Resource{ocpCaRotatedCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	kubeconfig := pulumi.All(getKC.Stdout, c.LBEIP.PublicIp).ApplyT(
		func(args []interface{}) string {
			return strings.ReplaceAll(args[0].(string),
				"https://api.crc.testing:6443",
				fmt.Sprintf("https://api.%s.nip.io:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kubeconfig, nil
}
