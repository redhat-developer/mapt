package kind

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/azure/modules/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/modules/virtual-machine"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/azure/services/network/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	utilKind "github.com/redhat-developer/mapt/pkg/targets/service/kind"
	"github.com/redhat-developer/mapt/pkg/util"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type kindRequest struct {
	mCtx              *mc.Context
	prefix            *string
	version           *string
	arch              *string
	spot              bool
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

func Create(mCtxArgs *mc.ContextArgs, args *utilKind.KindArgs) (*utilKind.KindResults, error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return nil, err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := &kindRequest{
		mCtx:              mCtx,
		prefix:            &prefix,
		version:           &args.Version,
		arch:              &args.Arch,
		extraPortMappings: args.ExtraPortMappings,
	}
	if args.Spot != nil {
		r.spot = args.Spot.Spot
	}
	ir, err := data.GetImageRef(data.Fedora, *r.arch, data.FedoraDefaultVersion)
	if err != nil {
		return nil, err
	}
	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			ComputeRequest: args.ComputeRequest,
			OSType:         "linux",
			ImageRef:       ir,
			Location:       &args.HostingPlace,
			Spot:           args.Spot})
	if err != nil {
		return nil, err
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackAzureKind),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
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

func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	// destroy
	return azure.Destroy(mCtx, stackAzureKind)
}

// Main function to deploy all requried resources to azure
func (r *kindRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKSpotPrice),
		pulumi.Float64(*r.allocationData.Price))
	// Get location for creating the Resource Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*r.allocationData.Location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureKindID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(r.mCtx.RunID()),
			Tags:              r.mCtx.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	// Extract hostPort values for LB target groups and security group rules
	extraHostPorts := make([]int, 0, len(r.extraPortMappings))
	for _, pm := range r.extraPortMappings {
		extraHostPorts = append(extraHostPorts, pm.HostPort)
	}
	sg, err := securityGroups(ctx, r.mCtx, r.prefix, r.allocationData.Location, extraHostPorts, rg)
	if err != nil {
		return err
	}
	n, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:        *r.prefix,
			ComponentID:   azureKindID,
			ResourceGroup: rg,
			Location:      r.allocationData.Location,
			SecurityGroup: sg,
		})
	if err != nil {
		return err
	}
	// Userdata
	udB64, err := userData(r.arch, r.version, r.extraPortMappings, n.PublicIP.IpAddress.Elem())
	if err != nil {
		return err
	}

	// Virutal machine
	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureKindID, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKPrivateKey),
		privateKey.PrivateKeyPem)
	vm, err := virtualmachine.Create(ctx, r.mCtx,
		&virtualmachine.VirtualMachineArgs{
			Prefix:          *r.prefix,
			ComponentID:     azureKindID,
			ResourceGroup:   rg,
			NetworkInteface: n.NetworkInterface,
			// Check this
			VMSize:           r.allocationData.ComputeSizes[0],
			Publisher:        r.allocationData.ImageRef.Publisher,
			Offer:            r.allocationData.ImageRef.Offer,
			Sku:              r.allocationData.ImageRef.Sku,
			ImageID:          r.allocationData.ImageRef.ID,
			PrivateKey:       privateKey,
			SpotPrice:        r.allocationData.Price,
			UserDataAsBase64: udB64,
			Location:         *r.allocationData.Location,
			AdminUsername:    amiUserDefault,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKUsername),
		pulumi.String(amiUserDefault))

	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKHost),
		n.PublicIP.IpAddress.Elem())

	kubeconfig, err := kubeconfig(ctx, r.prefix, n.PublicIP.IpAddress.Elem(), privateKey, []pulumi.Resource{vm})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, utilKind.OKKubeconfig),
		pulumi.ToSecret(kubeconfig))

	return nil
}

// security group for mac machine with ingress rules for ssh and vnc
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context,
	prefix, location *string, extraHostPorts []int,
	rg *resources.ResourceGroup) (securityGroup.SecurityGroup, error) {
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
	return securityGroup.Create(
		ctx,
		mCtx,
		&securityGroup.SecurityGroupArgs{
			Name:         resourcesUtil.GetResourceName(*prefix, azureKindID, "sg"),
			RG:           rg,
			Location:     location,
			IngressRules: ingressRules,
		})
}

func userData(arch, k8sVersion *string, parsedPortMappings []utilKind.PortMapping, ip pulumi.StringOutput) (pulumi.StringPtrInput, error) {
	ccB64 := ip.ApplyT(
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
	prefix *string, ip pulumi.StringOutput, mk *tls.PrivateKey,
	dependecies []pulumi.Resource,
) (pulumi.StringOutput, error) {
	// Once the cluster setup is comleted we
	// get the kubeconfig file from the host running the cluster
	// then we replace the internal access with the public IP
	// the resulting kubeconfig file can be used to access the cluster

	// Check cluster is ready
	kindReadyCmd, err := runCommand(ctx,
		command.CommandCloudInitWait,
		compute.LoggingCmdStd,
		fmt.Sprintf("%s-kind-readiness", *prefix), utilKind.KindID,
		mk, amiUserDefault, ip, dependecies)
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	// Get content for /opt/kubeconfig
	getKCCmd := ("cat /home/fedora/kubeconfig")
	getKC, err := runCommand(ctx,
		getKCCmd,
		compute.NoLoggingCmdStd,
		fmt.Sprintf("%s-kubeconfig", *prefix), utilKind.KindID, mk, amiUserDefault,
		ip, []pulumi.Resource{kindReadyCmd})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	kubeconfig := pulumi.All(getKC.Stdout, ip).ApplyT(
		func(args []interface{}) string {
			re := regexp.MustCompile(`https://[^:]+:\d+`)
			return re.ReplaceAllString(
				args[0].(string),
				fmt.Sprintf("https://%s:6443", args[1].(string)))
		}).(pulumi.StringOutput)
	return kubeconfig, nil
}

func runCommand(ctx *pulumi.Context,
	cmd string,
	loggingCmdStd bool,
	prefix, id string,
	mk *tls.PrivateKey, username string,
	ip pulumi.StringOutput,
	dependecies []pulumi.Resource) (*remote.Command, error) {
	return remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(prefix, id, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           ip,
				PrivateKey:     mk.PrivateKeyOpenssh,
				User:           pulumi.String(amiUserDefault),
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(cmd),
			Update: pulumi.String(cmd),
		},
		pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: "10m",
				Update: "10m"}),
		pulumi.DependsOn(dependecies))
}
