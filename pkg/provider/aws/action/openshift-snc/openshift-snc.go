package openshiftsnc

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
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
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type OpenshiftSNCArgs struct {
	Prefix         string
	ComputeRequest *cr.ComputeRequestArgs
	Version        string
	Arch           string
	PullSecretFile string
	Spot           *spotTypes.SpotArgs
	Timeout        string
}

type openshiftSNCRequest struct {
	mCtx           *mc.Context
	prefix         *string
	version        *string
	arch           *string
	spot           bool
	timeout        *string
	pullSecretFile *string
	allocationData *allocation.AllocationResult
}

func (r *openshiftSNCRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

type OpenshiftSncResultsMetadata struct {
	Username      string   `json:"username"`
	PrivateKey    string   `json:"private_key"`
	Host          string   `json:"host"`
	Kubeconfig    string   `json:"kubeconfig"`
	KubeadminPass string   `json:"kubeadmin_pass"`
	SpotPrice     *float64 `json:"spot_price,omitempty"`
	ConsoleUrl    string   `json:"console_url,omitempty"`
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(mCtxArgs *mc.ContextArgs, args *OpenshiftSNCArgs) (_ *OpenshiftSncResultsMetadata, err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return nil, err
	}
	// Compose request
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := openshiftSNCRequest{
		mCtx:           mCtx,
		prefix:         &prefix,
		version:        &args.Version,
		arch:           &args.Arch,
		pullSecretFile: &args.PullSecretFile,
		timeout:        &args.Timeout}
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
		return nil, err
	}
	// check if AMI exists
	amiName := amiName(&args.Version, &args.Arch)
	if err = checkAMIExists(&amiName, r.allocationData.Region, &args.Arch); err != nil {
		return nil, err
	}
	return r.createCluster()
}

// Will destroy resources related to machine
func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	logging.Debug("Run openshift destroy")
	// Create mapt Context
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	// Destroy fedora related resources
	if err = aws.DestroyStack(
		mCtx,
		aws.DestroyStackRequest{
			Stackname: stackName,
		}); err != nil {
		return err
	}
	// Destroy spot orchestrated stack
	if spot.Exist(mCtx) {
		return spot.Destroy(mCtx)
	}
	return nil
}

func (r *openshiftSNCRequest) createCluster() (*OpenshiftSncResultsMetadata, error) {
	cs := manager.Stack{
		StackName:   r.mCtx.StackNameByProject(stackName),
		ProjectName: r.mCtx.ProjectName(),
		BackedURL:   r.mCtx.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION:        *r.allocationData.Region,
				awsConstants.CONFIG_AWS_NATIVE_REGION: *r.allocationData.Region}),
		DeployFunc: r.deploy,
	}
	sr, err := manager.UpStack(r.mCtx, cs)
	if err != nil {
		return nil, fmt.Errorf("stack creation failed: %w", err)
	}

	return r.manageResults(sr, r.prefix)
}

func (r *openshiftSNCRequest) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		fmt.Sprintf("%s*", amiName(r.version, r.arch)),
		[]string{"self", amiOwner},
		map[string]string{
			"architecture": *r.arch})
	if err != nil {
		return err
	}
	nw, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:             *r.prefix,
			ID:                 awsOCPSNCID,
			Region:             *r.allocationData.Region,
			AZ:                 *r.allocationData.AZ,
			CreateLoadBalancer: r.allocationData.SpotPrice != nil,
			Airgap:             false,
		})
	if err != nil {
		return err
	}
	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(
			*r.prefix, awsOCPSNCID, "pk")}
	keyResources, err := kpr.Create(ctx, r.mCtx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)
	if r.mCtx.Debug() {
		keyResources.PrivateKey.PrivateKeyPem.ApplyT(
			func(privateKey string) (*string, error) {
				logging.Debugf("%s", privateKey)
				return nil, nil
			})
	}
	// Security groups
	securityGroups, err := securityGroups(ctx, r.mCtx, r.prefix, nw.Vpc)
	if err != nil {
		return err
	}
	// Instance profile required by logic within userdata
	iProfile, err := iam.InstanceProfile(ctx, r.prefix, &awsOCPSNCID, cloudConfigRequiredProfiles)
	if err != nil {
		return err
	}
	// Userdata
	udB64, kaPass, devPass, udDependecies, err := r.userData(ctx, &keyResources.PrivateKey.PublicKeyOpenssh, &nw.Eip.PublicIp)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeAdminPass),
		kaPass)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputDeveloperPass),
		devPass)
	// Create instance
	cr := compute.ComputeRequest{
		MCtx:             r.mCtx,
		Prefix:           *r.prefix,
		ID:               awsOCPSNCID,
		VPC:              nw.Vpc,
		Subnet:           nw.Subnet,
		AMI:              ami,
		KeyResources:     keyResources,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.InstanceTypes,
		DiskSize:         &diskSize,
		LB:               nw.LoadBalancer,
		Eip:              nw.Eip,
		LBTargetGroups:   []int{securityGroup.SSH_PORT, portHTTPS, portAPI},
		InstanceProfile:  iProfile,
		UserDataAsBase64: udB64,
		DependsOn:        udDependecies,
	}
	if r.spot && r.allocationData.SpotPrice != nil {
		cr.SpotPrice = *r.allocationData.SpotPrice
		cr.Spot = true
	}

	c, err := cr.NewCompute(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername),
		pulumi.String(amiUserDefault))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost),
		c.GetHostDnsName(true))
	if len(*r.timeout) > 0 {
		if err = serverless.OneTimeDelayedTask(ctx, r.mCtx,
			*r.allocationData.Region, *r.prefix,
			awsOCPSNCID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless",
				"openshift-snc",
				r.mCtx.ProjectName(),
				r.mCtx.BackedURL()),
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
func (r *openshiftSNCRequest) manageResults(stackResult auto.UpResult, prefix *string) (*OpenshiftSncResultsMetadata, error) {
	username, err := getResultOutput(outputUsername, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	privateKey, err := getResultOutput(outputUserPrivateKey, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	host, err := getResultOutput(outputHost, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	kubeconfig, err := getResultOutput(outputKubeconfig, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	kubeAdminPass, err := getResultOutput(outputKubeAdminPass, stackResult, prefix)
	if err != nil {
		return nil, err
	}

	hostIPKey := fmt.Sprintf("%s-%s", *prefix, outputHost)
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
		hostIPKey: "host",
		fmt.Sprintf("%s-%s", *prefix, outputKubeconfig):    "kubeconfig",
		fmt.Sprintf("%s-%s", *prefix, outputKubeAdminPass): "kubeadmin_pass",
		fmt.Sprintf("%s-%s", *prefix, outputDeveloperPass): "developer_pass",
	}

	outputPath := r.mCtx.GetResultsOutputPath()
	if len(outputPath) == 0 {
		logging.Warn("conn-details-output flag not set; skipping writing output files.")
	} else {
		if err := output.Write(stackResult, outputPath, results); err != nil {
			return nil, fmt.Errorf("failed to write results: %w", err)
		}
	}

	consoleURL := fmt.Sprintf(consoleURLRegex, host)
	if eip, ok := stackResult.Outputs[hostIPKey].Value.(string); ok {
		fmt.Printf("Cluster has been started you can access console at: %s.\n", fmt.Sprintf(consoleURLRegex, eip))
	}

	return &OpenshiftSncResultsMetadata{
		Username:      username,
		PrivateKey:    privateKey,
		Host:          host,
		Kubeconfig:    kubeconfig,
		KubeadminPass: kubeAdminPass,
		SpotPrice:     r.allocationData.SpotPrice,
		ConsoleUrl:    consoleURL,
	}, nil
}

// security group for Openshift
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix *string,
	vpc *ec2.Vpc) (pulumi.StringArray, error) {
	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:        resourcesUtil.GetResourceName(*prefix, awsOCPSNCID, "sg"),
		VPC:         vpc,
		Description: fmt.Sprintf("sg for %s", awsOCPSNCID),
		IngressRules: []securityGroup.IngressRules{securityGroup.SSH_TCP,
			{Description: "Console", FromPort: portHTTPS, ToPort: portHTTPS, Protocol: "tcp"},
			{Description: "API", FromPort: portAPI, ToPort: portAPI, Protocol: "tcp"}},
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

func checkAMIExists(amiName, region, arch *string) error {
	isAMIOffered, _, err := data.IsAMIOffered(
		data.ImageRequest{
			Name:   amiName,
			Arch:   arch,
			Region: region,
			Owner:  &amiOwner})
	if err != nil {
		return err
	}
	if !isAMIOffered {
		return fmt.Errorf("AMI %s could not be found in region: %s", *amiName, *region)
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

	// Check SSH connectivity first
	sshReadyCmd, err := c.RunCommand(ctx,
		command.CommandPing,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ssh-readiness", *prefix), awsOCPSNCID,
		mk, amiUserDefault, nil, c.Dependencies)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	// Check cluster is ready
	ocpReadyCmd, err := c.RunCommand(ctx,
		commandCrcReadiness,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-ocp-readiness", *prefix), awsOCPSNCID,
		mk, amiUserDefault, nil, []pulumi.Resource{sshReadyCmd})
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
	getKCCmd := ("sudo cat /opt/crc/kubeconfig")
	getKC, err := c.RunCommand(ctx,
		getKCCmd,
		compute.NoLoggingCmdStd,
		fmt.Sprintf("%s-kubeconfig", *prefix), awsOCPSNCID, mk, amiUserDefault,
		nil, []pulumi.Resource{ocpCaRotatedCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	kubeconfig := pulumi.All(getKC.Stdout, c.Eip.PublicIp).ApplyT(
		func(args []interface{}) string {
			return strings.ReplaceAll(args[0].(string),
				"https://api.crc.testing:6443",
				fmt.Sprintf("https://api.%s.nip.io:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kubeconfig, nil
}

func getResultOutput(name string, sr auto.UpResult, prefix *string) (string, error) {
	key := fmt.Sprintf("%s-%s", *prefix, name)
	output, ok := sr.Outputs[key]
	if !ok {
		return "", fmt.Errorf("output not found: %s", key)
	}
	value, ok := output.Value.(string)
	if !ok {
		return "", fmt.Errorf("output for %s is not a string", key)
	}
	return value, nil
}
