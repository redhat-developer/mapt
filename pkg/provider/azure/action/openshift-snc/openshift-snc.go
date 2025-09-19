package openshiftsnc

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	sncAPI "github.com/redhat-developer/mapt/pkg/provider/api/openshift-snc"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/modules/allocation"
	"github.com/redhat-developer/mapt/pkg/provider/azure/modules/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/modules/virtual-machine"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/azure/services/network/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

// move the cloud init template code to common openshift-snc api package
// parameterized it to modify the cloud-config based on the provider  to
// SSM for pull secret in AWS and Key Vault in case of Azure

const (
	stackAzureSNC = "stackAzureOpenshiftSNC"

	azureSNCId = "azsnc"

	outputHost           = "asncHost"
	outputUsername       = "asncUsername"
	outputUserPrivateKey = "asncUserPrivatekey"
	outputKubeadminPass  = "asncKubeadminPass"
	outputDeveloperPass  = "asncDeveloperPass"
	defaultVMSize        = "Standard_D8as_v5"
)

var (
	// snc
	defaultUsername = "core"
)

type openshiftSNCRequest struct {
	mCtx           *mc.Context `validate:"required"`
	prefix         *string
	arch           *string
	osType         *data.OSType
	version        *string
	username       *string
	pullSecretFile string
	allocationData *allocation.AllocationResult
}

//go:embed cloud-config
var CloudConfig []byte

func (r *openshiftSNCRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

func Create(mCtxArgs *mc.ContextArgs, args *sncAPI.OpenshiftSNCArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}

	osType := data.OpenShiftSNC
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := &openshiftSNCRequest{
		mCtx:           mCtx,
		prefix:         &prefix,
		arch:           &args.Arch,
		osType:         &osType,
		version:        &args.Version,
		username:       &defaultUsername,
		pullSecretFile: args.PullSecretFile,
	}

	ir, err := data.GetImageRef(*r.osType, *r.arch, *r.version)
	if err != nil {
		return err
	}

	r.allocationData, err = allocation.Allocation(mCtx,
		&allocation.AllocationArgs{
			ComputeRequest: args.ComputeRequest,
			OSType:         "linux",
			ImageRef:       ir,
			Location:       &args.Location,
			Spot:           args.Spot})
	if err != nil {
		return err
	}

	logging.Debug("Creating Linux Server")
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackAzureSNC),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, err := manager.UpStack(mCtx, cs)
	if err != nil {
		logging.Debugf("Error during upstack: %v", err)
	}
	return r.manageResults(sr)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	// destroy
	return azure.Destroy(mCtx, stackAzureSNC)
}

// Main function to deploy all requried resources to azure
func (r *openshiftSNCRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}

	// get suitable location for the resource group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*r.allocationData.Location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureSNCId, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(r.mCtx.RunID()),
			Tags:              r.mCtx.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	sg, err := securityGroups(ctx, r.mCtx, r.prefix, r.allocationData.Location, rg)
	if err != nil {
		return err
	}
	n, err := network.Create(ctx, r.mCtx, &network.NetworkArgs{
		Prefix:        *r.prefix,
		ComponentID:   azureSNCId,
		ResourceGroup: rg,
		Location:      r.allocationData.Location,
		SecurityGroup: sg,
	})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), n.PublicIP.IpAddress)
	// Virutal machine
	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureSNCId, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)

	// generate the snc cloud-config userdata
	userDataB64, kaPass, devPass, err := r.getUserData(ctx, n.PublicIP.IpAddress, privateKey.PublicKeyOpenssh)
	if err != nil {
		return fmt.Errorf("error creating RHEL Server on Azure: %v", err)
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputKubeadminPass), kaPass)
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputDeveloperPass), devPass)

	vmArgs := &virtualmachine.VirtualMachineArgs{
		Prefix:          *r.prefix,
		ComponentID:     azureSNCId,
		ResourceGroup:   rg,
		NetworkInteface: n.NetworkInterface,
		VMSize:          util.If(len(r.allocationData.ComputeSizes) > 0, r.allocationData.ComputeSizes[0], string(defaultVMSize)),
		ImageID:         r.allocationData.ImageRef.ID,
		AdminUsername:   *r.username,
		PrivateKey:      privateKey,
		SpotPrice:       r.allocationData.Price,
		Userdata:        userDataB64,
		DiskSizeGB:      256,
		Location:        *r.allocationData.Location,
	}
	vm, err := virtualmachine.Create(ctx, r.mCtx, vmArgs)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername), pulumi.String(*r.username))
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureSNCId, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           n.PublicIP.IpAddress.Elem(),
				PrivateKey:     privateKey.PrivateKeyOpenssh,
				User:           pulumi.String(*r.username),
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(command.CommandPing),
			Update: pulumi.String(command.CommandPing),
		},
		pulumi.Timeouts(
			&pulumi.CustomTimeouts{
				Create: "10m",
				Update: "10m"}),
		pulumi.DependsOn([]pulumi.Resource{vm}))
	return err
}

func Kubeconfig() {

}

// Write exported values in context to files o a selected target folder
func (r *openshiftSNCRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputUsername):       "username",
		fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey): "id_rsa",
		fmt.Sprintf("%s-%s", *r.prefix, outputHost):           "host",
		fmt.Sprintf("%s-%s", *r.prefix, outputKubeadminPass):  "kubeadmin_pass",
		fmt.Sprintf("%s-%s", *r.prefix, outputDeveloperPass):  "developer_pass",
	})
}

func (r *openshiftSNCRequest) getUserData(ctx *pulumi.Context, publicIP pulumi.StringPtrOutput, pubKey pulumi.StringOutput) (pulumi.StringOutput, pulumi.StringOutput, pulumi.StringOutput, error) {
	// KubeAdmin pass
	kaPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, azureSNCId, "kubeadminpassword"))
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}
	// Developer pass
	devPassword, err := security.CreatePassword(ctx,
		resourcesUtil.GetResourceName(
			*r.prefix, azureSNCId, "devpassword"))
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}
	// Manage pull secret
	ps, err := os.ReadFile(r.pullSecretFile)
	if err != nil {
		return pulumi.StringOutput{}, pulumi.StringOutput{}, pulumi.StringOutput{}, err
	}

	ccB64 := pulumi.All(pubKey, publicIP, kaPassword.Result, devPassword.Result).ApplyT(
		func(args []interface{}) (string, error) {
			var eip string
			ip, ok := args[1].(*string)
			if ok && ip != nil {
				eip = *ip
			}
			ccB64, err := sncAPI.GenCloudConfig(sncAPI.CloudConfigDataValues{
				Username:      defaultUsername,
				PubKey:        args[0].(string),
				PublicIP:      eip,
				PullSecret:    string(ps),
				PassKubeadmin: args[2].(string),
				PassDeveloper: args[3].(string),
			}, CloudConfig)
			return *ccB64, err
		}).(pulumi.StringOutput)

	return ccB64, kaPassword.Result, devPassword.Result, err
}

func securityGroups(ctx *pulumi.Context, mCtx *mc.Context, prefix, location *string, rg *resources.ResourceGroup) (securityGroup.SecurityGroup, error) {
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4

	consoleIngressRule := securityGroup.IngressRules{
		Description: "Console",
		FromPort:    sncAPI.PortHTTPS,
		ToPort:      sncAPI.PortHTTPS,
		Protocol:    "tcp",
	}

	apiSrvIngressRule := securityGroup.IngressRules{
		Description: "API",
		FromPort:    sncAPI.PortAPI,
		ToPort:      sncAPI.PortAPI,
		Protocol:    "tcp",
	}

	// Create SG with ingress rules
	return securityGroup.Create(
		ctx,
		mCtx,
		&securityGroup.SecurityGroupArgs{
			Name:     resourcesUtil.GetResourceName(*prefix, azureSNCId, "sg"),
			RG:       rg,
			Location: location,
			IngressRules: []securityGroup.IngressRules{
				sshIngressRule,
				consoleIngressRule,
				apiSrvIngressRule,
			},
		},
	)
}
