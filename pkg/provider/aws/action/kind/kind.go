package kind

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/serverless"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/spot"
	amiSVC "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/ami"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/keypair"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/aws/services/ec2/security-group"
	kindCloudConfig "github.com/redhat-developer/mapt/pkg/provider/util/cloud-config/kind"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type KindArgs struct {
	Prefix            string
	ComputeRequest    *cr.ComputeRequestArgs
	Version           string
	Arch              string
	Spot              *spotTypes.SpotArgs
	Timeout           string
	ExtraPortMappings []kindCloudConfig.PortMapping
}

type kindRequest struct {
	mCtx              *mc.Context
	prefix            *string
	version           *string
	arch              *string
	spot              bool
	timeout           *string
	allocationData    *allocation.AllocationResult
	extraPortMappings []kindCloudConfig.PortMapping
}

func (r *kindRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

type KindResultsMetadata struct {
	Username   string   `json:"username"`
	PrivateKey string   `json:"private_key"`
	Host       string   `json:"host"`
	Kubeconfig string   `json:"kubeconfig"`
	SpotPrice  *float64 `json:"spot_price,omitempty"`
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(mCtxArgs *mc.ContextArgs, args *KindArgs) (kr *KindResultsMetadata, err error) {
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return nil, err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := kindRequest{
		mCtx:              mCtx,
		prefix:            &prefix,
		version:           &args.Version,
		arch:              &args.Arch,
		timeout:           &args.Timeout,
		extraPortMappings: args.ExtraPortMappings}
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
	return r.createHost()
}

func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	logging.Debug("Run openshift destroy")
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	if err := aws.DestroyStack(mCtx, aws.DestroyStackRequest{Stackname: stackName}); err != nil {
		return err
	}
	if spot.Exist(mCtx) {
		return spot.Destroy(mCtx)
	}
	return nil
}

func (r *kindRequest) createHost() (*KindResultsMetadata, error) {
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

	metadataResults, err := r.manageResults(sr, r.prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to manage results: %w", err)
	}

	return metadataResults, nil
}

func (r *kindRequest) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		amiName(r.arch),
		[]string{amiOwner},
		map[string]string{"architecture": *r.arch},
		*r.allocationData.Region)
	if err != nil {
		return err
	}

	// Extract hostPort values for LB target groups and security group rules
	extraHostPorts := make([]int, 0, len(r.extraPortMappings))
	for _, pm := range r.extraPortMappings {
		extraHostPorts = append(extraHostPorts, pm.HostPort)
	}

	// Networking
	// LB is required if we use as which is used for spot feature
	nw, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:             *r.prefix,
			ID:                 awsKindID,
			Region:             *r.allocationData.Region,
			AZ:                 *r.allocationData.AZ,
			CreateLoadBalancer: r.allocationData.SpotPrice != nil,
		})
	if err != nil {
		return err
	}

	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(*r.prefix, awsKindID, "pk")}
	keyResources, err := kpr.Create(ctx, r.mCtx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey),
		keyResources.PrivateKey.PrivateKeyPem)

	// Security groups
	securityGroups, err := securityGroups(ctx, r.mCtx, r.prefix, nw.Vpc, extraHostPorts)
	if err != nil {
		return err
	}

	// Userdata
	udB64, err := userData(r.arch, r.version, r.extraPortMappings, &nw.Eip.PublicIp)
	if err != nil {
		return err
	}
	// Build LB target groups including both default and extra ports
	lbTargetGroups := []int{22, portAPI, portHTTP, portHTTPS}
	lbTargetGroups = append(lbTargetGroups, extraHostPorts...)

	cr := compute.ComputeRequest{
		MCtx:             r.mCtx,
		Prefix:           *r.prefix,
		ID:               awsKindID,
		VPC:              nw.Vpc,
		Subnet:           nw.Subnet,
		AMI:              ami,
		KeyResources:     keyResources,
		UserDataAsBase64: udB64,
		SecurityGroups:   securityGroups,
		InstaceTypes:     r.allocationData.InstanceTypes,
		DiskSize:         &diskSize,
		LB:               nw.LoadBalancer,
		Eip:              nw.Eip,
		LBTargetGroups:   lbTargetGroups,
	}
	if r.allocationData.SpotPrice != nil {
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
		c.GetHostIP(true))

	if len(*r.timeout) > 0 {
		err := serverless.OneTimeDelayedTask(ctx, r.mCtx,
			*r.allocationData.Region, *r.prefix,
			awsKindID,
			fmt.Sprintf("aws %s destroy --project-name %s --backed-url %s --serverless --force-destroy",
				"kind", r.mCtx.ProjectName(), r.mCtx.BackedURL()),
			*r.timeout)
		if err != nil {
			return err
		}
	}

	kubeconfig, err := kubeconfig(ctx, r.prefix, c, keyResources.PrivateKey)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeconfig),
		pulumi.ToSecret(kubeconfig))

	return nil
}

// Write exported values in context to files o a selected target folder
func (r *kindRequest) manageResults(stackResult auto.UpResult, prefix *string) (*KindResultsMetadata, error) {
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

	metadataResults := &KindResultsMetadata{
		Username:   username,
		PrivateKey: privateKey,
		Host:       host,
		Kubeconfig: kubeconfig,
		SpotPrice:  r.allocationData.SpotPrice,
	}

	hostIPKey := fmt.Sprintf("%s-%s", *prefix, outputHost)
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, outputUserPrivateKey): "id_rsa",
		hostIPKey: "host",
		fmt.Sprintf("%s-%s", *prefix, outputKubeconfig): "kubeconfig",
	}

	if r.mCtx.GetResultsOutputPath() != "" {
		if err := output.Write(stackResult, r.mCtx.GetResultsOutputPath(), results); err != nil {
			return nil, fmt.Errorf("failed to write results: %w", err)
		}
	}

	return metadataResults, nil
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

// security group for Openshift
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix *string,
	vpc *ec2.Vpc, extraHostPorts []int) (pulumi.StringArray, error) {
	// Build ingress rules including both default and extra ports
	ingressRules := []securityGroup.IngressRules{
		securityGroup.SSH_TCP,
		{Description: "HTTPS", FromPort: portHTTPS, ToPort: portHTTPS, Protocol: "tcp"},
		{Description: "HTTP", FromPort: portHTTP, ToPort: portHTTP, Protocol: "tcp"},
		{Description: "API", FromPort: portAPI, ToPort: portAPI, Protocol: "tcp"},
	}

	// Add extra ports to ingress rules
	for _, port := range extraHostPorts {
		ingressRules = append(ingressRules, securityGroup.IngressRules{
			Description: fmt.Sprintf("Extra Port %d", port),
			FromPort:    port,
			ToPort:      port,
			Protocol:    "tcp",
		})
	}

	// Create SG with ingress rules
	sg, err := securityGroup.SGRequest{
		Name:         resourcesUtil.GetResourceName(*prefix, awsKindID, "sg"),
		VPC:          vpc,
		Description:  fmt.Sprintf("sg for %s", awsKindID),
		IngressRules: ingressRules,
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

func userData(arch, k8sVersion *string, parsedPortMappings []kindCloudConfig.PortMapping, lbEIP *pulumi.StringOutput) (pulumi.StringPtrInput, error) {
	ccB64 := lbEIP.ApplyT(
		func(publicIP string) (string, error) {
			ccB64, err := kindCloudConfig.CloudConfig(
				&kindCloudConfig.DataValues{
					Arch: util.If(*arch == "x86_64",
						kindCloudConfig.X86_64,
						kindCloudConfig.Arm64),
					KindVersion:       KindK8sVersions[*k8sVersion].kindVersion,
					KindImage:         KindK8sVersions[*k8sVersion].KindImage,
					Username:          amiUserDefault,
					PublicIP:          publicIP,
					ExtraPortMappings: parsedPortMappings})
			return *ccB64, err
		}).(pulumi.StringOutput)

	return ccB64, nil
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
	kindReadyCmd, err := c.RunCommand(ctx,
		command.CommandCloudInitWait,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-kind-readiness", *prefix), awsKindID,
		mk, amiUserDefault, nil, c.Dependencies)
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	// Get content for /opt/kubeconfig
	getKCCmd := ("cat /home/fedora/kubeconfig")
	getKC, err := c.RunCommand(ctx,
		getKCCmd,
		compute.NoLoggingCmdStd,
		fmt.Sprintf("%s-kubeconfig", *prefix), awsKindID, mk, amiUserDefault,
		nil, []pulumi.Resource{kindReadyCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	kubeconfig := pulumi.All(getKC.Stdout, c.Eip.PublicIp).ApplyT(
		func(args []interface{}) string {
			re := regexp.MustCompile(`https://[^:]+:\d+`)
			return re.ReplaceAllString(
				args[0].(string),
				fmt.Sprintf("https://%s:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kubeconfig, nil
}
