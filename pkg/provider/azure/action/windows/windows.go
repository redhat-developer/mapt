package windows

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/storage/v2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/azure"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/azure/module/network"
	virtualmachine "github.com/redhat-developer/mapt/pkg/provider/azure/module/virtual-machine"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/file"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
	"golang.org/x/exp/slices"
)

//go:embed rhqp-ci-setup.ps1
var RHQPCISetupScript []byte

type WindowsRequest struct {
	Prefix              string
	Location            string
	VMSizes             []string
	InstaceTypeRequest  instancetypes.InstanceRequest
	Version             string
	Feature             string
	Username            string
	AdminUsername       string
	Spot                bool
	SpotTolerance       spotAzure.EvictionRate
	SpotExcludedRegions []string
	Profiles            []string
}

type ghActionsRunnerData struct {
	ActionsRunnerSnippet string
	CirrusSnippet        string
}

func Create(ctx *maptContext.ContextArgs, r *WindowsRequest) (err error) {
	// Create mapt Context
	if err := maptContext.Init(ctx, azure.Provider()); err != nil {
		return err
	}

	if len(r.VMSizes) == 0 {
		vmSizes, err := r.InstaceTypeRequest.GetMachineTypes()
		if err != nil {
			logging.Debugf("Failed to get instance types: %v", err)
		}
		if len(vmSizes) > 0 {
			r.VMSizes = append(r.VMSizes, vmSizes...)
		}
	}
	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackCreateWindowsDesktop),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, _ := manager.UpStack(cs)
	return r.manageResults(sr)
}

func Destroy(ctx *maptContext.ContextArgs) error {
	// Create mapt Context
	if err := maptContext.Init(ctx, azure.Provider()); err != nil {
		return err
	}
	// destroy
	return azure.Destroy(
		maptContext.ProjectName(),
		maptContext.BackedURL(),
		maptContext.StackNameByProject(stackCreateWindowsDesktop))
}

// Main function to deploy all requried resources to azure
func (r *WindowsRequest) deployer(ctx *pulumi.Context) error {
	logging.Debugf("Using these VM types for Spot price query: %v", r.VMSizes)
	// Get values for spot machine
	location, vmType, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}
	// Get location for creating Resource Group
	rgLocation := azure.GetSuitableLocationForResourceGroup(*location)
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(rgLocation),
			ResourceGroupName: pulumi.String(maptContext.RunID()),
			Tags:              maptContext.ResourceTags(),
		})
	if err != nil {
		return err
	}
	// Networking
	nr := network.NetworkRequest{
		Prefix:        r.Prefix,
		ComponentID:   azureWindowsDesktopID,
		ResourceGroup: rg,
	}
	n, err := nr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost), n.PublicIP.IpAddress)
	// Virutal machine
	// TODO check if validation should be moved to the top of the func?
	if err := r.validateProfiles(); err != nil {
		return err
	}
	adminPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pswd-adminuser"))
	if err != nil {
		return err
	}
	vmr := virtualmachine.VirtualMachineRequest{
		Prefix:          r.Prefix,
		ComponentID:     azureWindowsDesktopID,
		ResourceGroup:   rg,
		NetworkInteface: n.NetworkInterface,
		VMSize:          vmType,
		Publisher:       "MicrosoftWindowsDesktop",
		Offer:           fmt.Sprintf("windows-%s", r.Version),
		Sku:             fmt.Sprintf("win%s-%s", r.Version, r.Feature),
		AdminUsername:   r.AdminUsername,
		AdminPasswd:     adminPasswd,
		SpotPrice:       spotPrice,
		Location:        rgLocation,
	}
	vm, err := vmr.Create(ctx)
	if err != nil {
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputAdminUsername), pulumi.String(r.AdminUsername))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputAdminUserPassword), adminPasswd.Result)
	// Setup machine on post init (may move too to virtual-machine pkg)
	pk, vme, err := r.postInitSetup(ctx, rg, vm, *location)
	if err != nil {
		return err
	}
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           n.PublicIP.IpAddress.Elem(),
				PrivateKey:     pk.PrivateKeyOpenssh,
				User:           pulumi.String(r.Username),
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

func (r *WindowsRequest) valuesCheckingSpot() (*string, string, *float64, error) {
	if r.Spot {
		bsc, err :=
			spotAzure.GetBestSpotChoice(spotAzure.BestSpotChoiceRequest{
				VMTypes:               util.If(len(r.VMSizes) > 0, r.VMSizes, []string{defaultVMSize}),
				OSType:                "windows",
				EvictionRateTolerance: r.SpotTolerance,
				ExcludedRegions:       r.SpotExcludedRegions,
			})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, "", nil, err
		}
		return &bsc.Location, bsc.VMType, &bsc.Price, nil
	}
	// TODO we need to extend this to other azure targets (refactor this function)
	// plus we probably would need to check prices for vmsizes and pick the cheaper
	availableVMSizes, err := data.FilterVMSizeOfferedByLocation(r.VMSizes, r.Location)
	if err != nil {
		return nil, "", nil, err
	}
	if len(availableVMSizes) == 0 {
		return nil, "", nil, fmt.Errorf("no vm size mathing expectations on current region")
	}
	return &r.Location, availableVMSizes[0], nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *WindowsRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, maptContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputAdminUsername):     "adminusername",
		fmt.Sprintf("%s-%s", r.Prefix, outputAdminUserPassword): "adminuserpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):          "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword):      "userpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey):    "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):              "host",
	})
}

// run a post script to setup the machine as expected according to rhqp-ci-setup.ps1
// it also exports to pulumi context user name, user password and user privatekey
func (r *WindowsRequest) postInitSetup(ctx *pulumi.Context, rg *resources.ResourceGroup,
	vm *compute.VirtualMachine, location string) (*tls.PrivateKey, *compute.VirtualMachineExtension, error) {
	userPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pswd-user"))
	if err != nil {
		return nil, nil, err

	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUsername), pulumi.String(r.Username))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword), userPasswd.Result)
	privateKey, err := tls.NewPrivateKey(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "privatekey-user"),
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return nil, nil, err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)
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
				r.Username,
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
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "ext"),
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
			Tags: maptContext.ResourceTags(),
		},
		pulumi.RetainOnDelete(true))
	return privateKey, vme, err
}

// Upload scrip to blob container to be used within Microsoft Compute extension
func (r *WindowsRequest) uploadScript(ctx *pulumi.Context,
	rg *resources.ResourceGroup, location string) (*storage.Blob, error) {
	sa, err := storage.NewStorageAccount(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "sa"),
		&storage.StorageAccountArgs{
			AccountName:           pulumi.String(maptContext.RunID()),
			Kind:                  pulumi.String("BlockBlobStorage"),
			ResourceGroupName:     rg.Name,
			Location:              pulumi.String(location),
			AllowBlobPublicAccess: pulumi.BoolPtr(true),
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Premium_LRS"),
			},
			Tags: maptContext.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	c, err := storage.NewBlobContainer(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "co"),
		&storage.BlobContainerArgs{
			ContainerName:     pulumi.String(maptContext.RunID()),
			AccountName:       sa.Name,
			ResourceGroupName: rg.Name,
			PublicAccess:      storage.PublicAccessBlob,
		})
	if err != nil {
		return nil, err
	}
	cirrusSnippet, err := integrations.GetIntegrationSnippet(cirrus.GetRunnerArgs(), r.Username)
	if err != nil {
		return nil, err
	}
	ghActionsRunnerSnippet, err := integrations.GetIntegrationSnippet(github.GetRunnerArgs(), r.Username)
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
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "bl"),
		&storage.BlobArgs{
			AccountName:       sa.Name,
			ContainerName:     c.Name,
			ResourceGroupName: rg.Name,
			Source:            pulumi.NewStringAsset(ciSetupScript),
			BlobName:          pulumi.String(scriptName),
		})
}

// Check if profiles for the target hosts are supported
func (r *WindowsRequest) validateProfiles() error {
	for _, p := range r.Profiles {
		if !slices.Contains(profiles, p) {
			return fmt.Errorf("the profile %s is not supported", p)
		}
	}
	return nil
}

// Check if a request contains a profile
func (r *WindowsRequest) profilesAsParams() string {
	pp := util.ArrayConvert(
		r.Profiles,
		func(p string) string {
			return fmt.Sprintf("-%sProfile", p)
		})
	return strings.Join(pp, " ")
}
