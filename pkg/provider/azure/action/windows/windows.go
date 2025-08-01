package windows

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-azure-native-sdk/compute/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/storage/v3"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	infra "github.com/redhat-developer/mapt/pkg/provider"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/module/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/module/virtual-machine"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/azure/services/network/security-group"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
	"golang.org/x/exp/slices"
)

//go:embed rhqp-ci-setup.ps1
var RHQPCISetupScript []byte

type WindowsArgs struct {
	Prefix              string
	Location            string
	ComputeRequest      *cr.ComputeRequestArgs
	Version             string
	Feature             string
	Username            string
	AdminUsername       string
	Spot                bool
	SpotTolerance       spotTypes.Tolerance
	SpotExcludedRegions []string
	Profiles            []string
}

type windowsRequest struct {
	mCtx                *mc.Context `validate:"required"`
	prefix              *string
	location            *string
	vmSizes             []string
	version             *string
	feature             *string
	username            *string
	adminUsername       *string
	spot                *bool
	spotTolerance       *spotTypes.Tolerance
	spotExcludedRegions []string
	profiles            []string
}

func (r *windowsRequest) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

type ghActionsRunnerData struct {
	ActionsRunnerSnippet string
	CirrusSnippet        string
}

func Create(mCtxArgs *mc.ContextArgs, args *WindowsArgs) (err error) {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	prefix := util.If(len(args.Prefix) > 0, args.Prefix, "main")
	r := &windowsRequest{
		mCtx:                mCtx,
		prefix:              &prefix,
		location:            &args.Location,
		version:             &args.Version,
		username:            &args.Username,
		spot:                &args.Spot,
		spotTolerance:       &args.SpotTolerance,
		spotExcludedRegions: args.SpotExcludedRegions,
		feature:             &args.Feature,
		adminUsername:       &args.AdminUsername,
		profiles:            args.Profiles,
	}
	if len(args.ComputeRequest.ComputeSizes) > 0 {
		r.vmSizes = args.ComputeRequest.ComputeSizes
	} else {
		vmSizes, err :=
			data.NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return err
		}
		r.vmSizes = vmSizes
	}
	cs := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackCreateWindowsDesktop),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, _ := manager.UpStack(mCtx, cs)
	return r.manageResults(sr)
}

func Destroy(mCtxArgs *mc.ContextArgs) error {
	// Create mapt Context
	mCtx, err := mc.Init(mCtxArgs, azure.Provider())
	if err != nil {
		return err
	}
	// destroy
	return azure.Destroy(mCtx, stackCreateWindowsDesktop)
}

// Main function to deploy all requried resources to azure
func (r *windowsRequest) deployer(ctx *pulumi.Context) error {
	if err := r.validate(); err != nil {
		return err
	}
	logging.Debugf("Using these VM types for Spot price query: %v", r.vmSizes)
	// Get values for spot machine
	location, vmType, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}
	// Get location for creating Resource Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(r.mCtx.RunID()),
			Tags:              r.mCtx.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	sg, err := securityGroups(ctx, r.mCtx, r.prefix, location, rg)
	if err != nil {
		return err
	}
	n, err := network.Create(ctx, r.mCtx,
		&network.NetworkArgs{
			Prefix:        *r.prefix,
			ComponentID:   azureWindowsDesktopID,
			ResourceGroup: rg,
			Location:      location,
			SecurityGroup: sg,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputHost), n.PublicIP.IpAddress)
	// Virutal machine
	// TODO check if validation should be moved to the top of the func?
	if err := r.validateProfiles(); err != nil {
		return err
	}
	adminPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "pswd-adminuser"))
	if err != nil {
		return err
	}
	vm, err := virtualmachine.Create(ctx, r.mCtx,
		&virtualmachine.VirtualMachineArgs{
			Prefix:          *r.prefix,
			ComponentID:     azureWindowsDesktopID,
			ResourceGroup:   rg,
			NetworkInteface: n.NetworkInterface,
			VMSize:          vmType,
			Publisher:       "MicrosoftWindowsDesktop",
			Offer:           fmt.Sprintf("windows-%s", *r.version),
			Sku:             fmt.Sprintf("win%s-%s", *r.version, *r.feature),
			AdminUsername:   *r.adminUsername,
			AdminPasswd:     adminPasswd,
			SpotPrice:       spotPrice,
			Location:        *location,
		})
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputAdminUsername), pulumi.String(*r.adminUsername))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputAdminUserPassword), adminPasswd.Result)
	// Setup machine on post init (may move too to virtual-machine pkg)
	pk, vme, err := r.postInitSetup(ctx, rg, vm, *location)
	if err != nil {
		return err
	}
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           n.PublicIP.IpAddress.Elem(),
				PrivateKey:     pk.PrivateKeyOpenssh,
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
		pulumi.DependsOn([]pulumi.Resource{vme}))
	return err
}

// security group for mac machine with ingress rules for ssh and vnc
func securityGroups(ctx *pulumi.Context, mCtx *mc.Context,
	prefix, location *string,
	rg *resources.ResourceGroup) (securityGroup.SecurityGroup, error) {
	// ingress for ssh access from 0.0.0.0
	sshIngressRule := securityGroup.SSH_TCP
	sshIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	rdpIngressRule := securityGroup.RDP_TCP
	rdpIngressRule.CidrBlocks = infra.NETWORKING_CIDR_ANY_IPV4
	// Create SG with ingress rules
	return securityGroup.Create(
		ctx,
		mCtx,
		&securityGroup.SecurityGroupArgs{
			Name:     resourcesUtil.GetResourceName(*prefix, azureWindowsDesktopID, "sg"),
			RG:       rg,
			Location: location,
			IngressRules: []securityGroup.IngressRules{
				sshIngressRule, rdpIngressRule},
		})
}

func (r *windowsRequest) valuesCheckingSpot() (*string, string, *float64, error) {
	if *r.spot {
		bsc, err :=
			data.SpotInfo(&data.SpotInfoArgs{
				ComputeSizes: util.If(len(r.vmSizes) > 0, r.vmSizes, []string{defaultVMSize}),
				OSType:       "windows",
				// EvictionRateTolerance: r.SpotTolerance,
				ExcludedLocations: r.spotExcludedRegions,
			})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, "", nil, err
		}
		return &bsc.HostingPlace, bsc.ComputeType, &bsc.Price, nil
	}
	// TODO we need to extend this to other azure targets (refactor this function)
	// plus we probably would need to check prices for vmsizes and pick the cheaper
	availableVMSizes, err := data.FilterVMSizeOfferedByLocation(r.vmSizes, *r.location)
	if err != nil {
		return nil, "", nil, err
	}
	if len(availableVMSizes) == 0 {
		return nil, "", nil, fmt.Errorf("no vm size mathing expectations on current region")
	}
	return r.location, availableVMSizes[0], nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *windowsRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, r.mCtx.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", *r.prefix, outputAdminUsername):     "adminusername",
		fmt.Sprintf("%s-%s", *r.prefix, outputAdminUserPassword): "adminuserpassword",
		fmt.Sprintf("%s-%s", *r.prefix, outputUsername):          "username",
		fmt.Sprintf("%s-%s", *r.prefix, outputUserPassword):      "userpassword",
		fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey):    "id_rsa",
		fmt.Sprintf("%s-%s", *r.prefix, outputHost):              "host",
	})
}

// run a post script to setup the machine as expected according to rhqp-ci-setup.ps1
// it also exports to pulumi context user name, user password and user privatekey
func (r *windowsRequest) postInitSetup(ctx *pulumi.Context, rg *resources.ResourceGroup,
	vm *compute.VirtualMachine, location string) (*tls.PrivateKey, *compute.VirtualMachineExtension, error) {
	userPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "pswd-user"))
	if err != nil {
		return nil, nil, err

	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUsername), pulumi.String(*r.username))
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPassword), userPasswd.Result)
	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return nil, nil, err
	}
	ctx.Export(fmt.Sprintf("%s-%s", *r.prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)
	// upload the script to a ephemeral blob container
	b, err := r.uploadScript(ctx, rg, location)
	if err != nil {
		return nil, nil, err
	}
	// the post script command will be generated based on generated data as parameters
	setupCommand := pulumi.All(userPasswd.Result, privateKey.PublicKeyOpenssh, vm.OsProfile.ComputerName()).ApplyT(
		func(args []interface{}) string {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			hostname := args[2].(*string)
			return fmt.Sprintf(
				"powershell -ExecutionPolicy Unrestricted -File %s %s -userPass \"%s\" -user %s -hostname %s -ghToken \"%s\" -cirrusToken \"%s\" -authorizedKey \"%s\"",
				scriptName,
				r.profilesAsParams(),
				password,
				*r.username,
				*hostname,
				github.GetToken(),
				cirrus.GetToken(),
				authorizedKey,
			)
		}).(pulumi.StringOutput)
	// the post script will be executed as a extension,
	// this resource is retain on delete b/c it does not create a real resource on the provider
	// and also if vm where it has been executed is stopped (i.e. deallocated spot instance) it can
	// not be deleted leading to break all destroy operation on the resources.
	vme, err := compute.NewVirtualMachineExtension(
		ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "ext"),
		&compute.VirtualMachineExtensionArgs{
			ResourceGroupName:  rg.Name,
			Location:           pulumi.String(location),
			VmName:             vm.Name,
			Publisher:          pulumi.String("Microsoft.Compute"),
			Type:               pulumi.String("CustomScriptExtension"),
			TypeHandlerVersion: pulumi.String("1.10"),
			ProtectedSettings: pulumi.Map{
				"fileUris": pulumi.Array{
					b.Url,
				},
				"commandToExecute": setupCommand,
			},
			Tags: r.mCtx.ResourceTags(),
		},
		pulumi.RetainOnDelete(true))
	return privateKey, vme, err
}

// Upload scrip to blob container to be used within Microsoft Compute extension
func (r *windowsRequest) uploadScript(ctx *pulumi.Context,
	rg *resources.ResourceGroup, location string) (*storage.Blob, error) {
	sa, err := storage.NewStorageAccount(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "sa"),
		&storage.StorageAccountArgs{
			AccountName:           pulumi.String(r.mCtx.RunID()),
			Kind:                  pulumi.String("BlockBlobStorage"),
			ResourceGroupName:     rg.Name,
			Location:              pulumi.String(location),
			AllowBlobPublicAccess: pulumi.BoolPtr(true),
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Premium_LRS"),
			},
			Tags: r.mCtx.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	c, err := storage.NewBlobContainer(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "co"),
		&storage.BlobContainerArgs{
			ContainerName:     pulumi.String(r.mCtx.RunID()),
			AccountName:       sa.Name,
			ResourceGroupName: rg.Name,
			PublicAccess:      storage.PublicAccessBlob,
		})
	if err != nil {
		return nil, err
	}
	cirrusSnippet, err := integrations.GetIntegrationSnippet(cirrus.GetRunnerArgs(), *r.username)
	if err != nil {
		return nil, err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippet(github.GetRunnerArgs(), *r.username)
	if err != nil {
		return nil, err
	}
	logging.Debug("got the self hosted runner script")
	ciSetupScript, err := file.Template(
		ghActionsRunnerData{
			*ghActionsRunnerSnippet,
			*cirrusSnippet,
		},
		string(RHQPCISetupScript))
	if err != nil {
		return nil, err
	}

	return storage.NewBlob(ctx,
		resourcesUtil.GetResourceName(*r.prefix, azureWindowsDesktopID, "bl"),
		&storage.BlobArgs{
			AccountName:       sa.Name,
			ContainerName:     c.Name,
			ResourceGroupName: rg.Name,
			Source:            pulumi.NewStringAsset(ciSetupScript),
			BlobName:          pulumi.String(scriptName),
		})
}

// Check if profiles for the target hosts are supported
func (r *windowsRequest) validateProfiles() error {
	for _, p := range r.profiles {
		if !slices.Contains(profiles, p) {
			return fmt.Errorf("the profile %s is not supported", p)
		}
	}
	return nil
}

// Check if a request contains a profile
func (r *windowsRequest) profilesAsParams() string {
	pp := util.ArrayConvert(
		r.profiles,
		func(p string) string {
			return fmt.Sprintf("-%sProfile", p)
		})
	return strings.Join(pp, " ")
}
