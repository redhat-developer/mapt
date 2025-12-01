package kind

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
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
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	utilKind "github.com/redhat-developer/mapt/pkg/targets/service/kind"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type kindRequest struct {
	mCtx              *mc.Context
	prefix            *string
	version           *string
	arch              *string
	spot              bool
	timeout           *string
	allocationData    *allocation.AllocationResult
	extraPortMappings []utilKind.PortMapping
}

func (r *kindRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

// Create orchestrate 3 stacks:
// If spot is enable it will run best spot option to get the best option to spin the machine
// Then it will run the stack for windows dedicated host
func Create(mCtxArgs *mc.ContextArgs, args *utilKind.KindArgs) (kr *utilKind.KindResults, err error) {
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
	if err := aws.DestroyStack(mCtx, aws.DestroyStackRequest{Stackname: utilKind.StackName}); err != nil {
		return err
	}
	if spot.Exist(mCtx) {
		if err := spot.Destroy(mCtx); err != nil {
			return err
		}
	}

	// Cleanup S3 state after all stacks have been destroyed
	return aws.CleanupState(mCtx)
}

func (r *kindRequest) createHost() (*utilKind.KindResults, error) {
	cs := manager.Stack{
		StackName:   r.mCtx.StackNameByProject(utilKind.StackName),
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

	metadataResults, err := utilKind.Results(r.mCtx, sr, r.prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to manage results: %w", err)
	}

	return metadataResults, nil
}

func (r *kindRequest) deploy(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKSpotPrice),
		pulumi.Float64(*r.allocationData.SpotPrice))
	// Get AMI
	ami, err := amiSVC.GetAMIByName(ctx,
		amiName(r.arch),
		[]string{amiOwner},
		map[string]string{"architecture": *r.arch})
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
			ID:                 utilKind.KindID,
			Region:             *r.allocationData.Region,
			AZ:                 *r.allocationData.AZ,
			CreateLoadBalancer: r.allocationData.SpotPrice != nil,
		})
	if err != nil {
		return err
	}

	// Create Keypair
	kpr := keypair.KeyPairRequest{
		Name: resourcesUtil.GetResourceName(*r.prefix, utilKind.KindID, "pk")}
	keyResources, err := kpr.Create(ctx, r.mCtx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKPrivateKey),
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
	lbTargetGroups := []int{22, utilKind.PortAPI, utilKind.PortHTTP, utilKind.PortHTTPS}
	lbTargetGroups = append(lbTargetGroups, extraHostPorts...)

	cr := compute.ComputeRequest{
		MCtx:             r.mCtx,
		Prefix:           *r.prefix,
		ID:               utilKind.KindID,
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

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKUsername),
		pulumi.String(amiUserDefault))

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKHost),
		c.GetHostIP(true))

	if len(*r.timeout) > 0 {
		err := serverless.OneTimeDelayedTask(ctx, r.mCtx,
			*r.allocationData.Region, *r.prefix,
			utilKind.KindID,
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
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKKubeconfig),
		pulumi.ToSecret(kubeconfig))

	return nil
}

// security group for Openshift
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix *string,
	vpc *ec2.Vpc, extraHostPorts []int) (pulumi.StringArray, error) {
	// Build ingress rules including both default and extra ports
	ingressRules := []securityGroup.IngressRules{
		securityGroup.SSH_TCP,
		{Description: "HTTPS", FromPort: utilKind.PortHTTPS, ToPort: utilKind.PortHTTPS, Protocol: "tcp"},
		{Description: "HTTP", FromPort: utilKind.PortHTTP, ToPort: utilKind.PortHTTP, Protocol: "tcp"},
		{Description: "API", FromPort: utilKind.PortAPI, ToPort: utilKind.PortAPI, Protocol: "tcp"},
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
		Name:         resourcesUtil.GetResourceName(*prefix, utilKind.KindID, "sg"),
		VPC:          vpc,
		Description:  fmt.Sprintf("sg for %s", utilKind.KindID),
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

func userData(arch, k8sVersion *string, parsedPortMappings []utilKind.PortMapping, lbEIP *pulumi.StringOutput) (pulumi.StringPtrInput, error) {
	ccB64 := lbEIP.ApplyT(
		func(publicIP string) (string, error) {
			cc := &utilKind.CloudConfigArgs{
				Arch: util.If(*arch == "x86_64",
					utilKind.X86_64,
					utilKind.Arm64),
				KindVersion:       utilKind.KindK8sVersions[*k8sVersion].KindVersion,
				KindImage:         utilKind.KindK8sVersions[*k8sVersion].KindImage,
				Username:          amiUserDefault,
				PublicIP:          publicIP,
				ExtraPortMappings: parsedPortMappings}
			ccB64, err := cc.CloudConfig()
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
		fmt.Sprintf("%s-kind-readiness", *prefix), utilKind.KindID,
		mk, amiUserDefault, nil, c.Dependencies)
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	// Get content for /opt/kubeconfig
	getKCCmd := ("cat /home/fedora/kubeconfig")
	getKC, err := c.RunCommand(ctx,
		getKCCmd,
		compute.NoLoggingCmdStd,
		fmt.Sprintf("%s-kubeconfig", *prefix), utilKind.KindID, mk, amiUserDefault,
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
