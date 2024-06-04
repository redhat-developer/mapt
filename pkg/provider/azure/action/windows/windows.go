package windows

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/azure"
	spotprice "github.com/adrianriobo/qenvs/pkg/provider/azure/module/spot-price"
	"github.com/adrianriobo/qenvs/pkg/provider/util/command"
	"github.com/adrianriobo/qenvs/pkg/provider/util/output"
	"github.com/adrianriobo/qenvs/pkg/provider/util/security"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/storage/v2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/exp/slices"
)

//go:embed rhqp-ci-setup.ps1
var RHQPCISetupScript []byte

type WindowsRequest struct {
	Prefix        string
	Location      string
	VMSize        string
	Version       string
	Feature       string
	Username      string
	AdminUsername string
	Spot          bool
	SpotTolerance spotprice.EvictionRate
	Profiles      []string
}

func Create(r *WindowsRequest) (err error) {
	logging.Debug("Creating Windows Desktop")
	cs := manager.Stack{
		StackName:           qenvsContext.StackNameByProject(stackCreateWindowsDesktop),
		ProjectName:         qenvsContext.ProjectName(),
		BackedURL:           qenvsContext.BackedURL(),
		ProviderCredentials: azure.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	sr, _ := manager.UpStack(cs)
	return r.manageResults(sr)
}

func Destroy() error {
	if err := azure.Destroy(
		qenvsContext.ProjectName(),
		qenvsContext.BackedURL(),
		qenvsContext.StackNameByProject(stackCreateWindowsDesktop)); err != nil {
		return err
	}
	return azure.Destroy(
		qenvsContext.ProjectName(),
		qenvsContext.BackedURL(),
		qenvsContext.StackNameByProject(stackSyncWindowsDesktop))
}

// Main function to deploy all requried resources to azure
func (r *WindowsRequest) deployer(ctx *pulumi.Context) error {
	location, spotPrice, err := r.valuesCheckingSpot()
	if err != nil {
		return err
	}
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(*location),
			ResourceGroupName: pulumi.String(qenvsContext.RunID()),
			Tags:              qenvsContext.ResourceTags(),
		})
	if err != nil {
		return err
	}
	ni, pi, err := r.createNetworking(ctx, rg, *location)
	if err != nil {
		return err
	}
	vm, err := r.createVirtualMachine(ctx, rg, ni, *location, spotPrice)
	if err != nil {
		return err
	}
	pk, vme, err := r.postInitSetup(ctx, rg, vm, *location)
	if err != nil {
		return err
	}
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           pi.IpAddress.Elem(),
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

func (r *WindowsRequest) valuesCheckingSpot() (*string, *float64, error) {
	if r.Spot {
		bsc, err :=
			spotprice.GetBestSpotChoice(spotprice.BestSpotChoiceRequest{
				VMTypes:              []string{r.VMSize},
				OSType:               "windows",
				EvictioRateTolerance: r.SpotTolerance,
			})
		logging.Debugf("Best spot price option found: %v", bsc)
		if err != nil {
			return nil, nil, err
		}
		return &bsc.Location, &bsc.Price, nil
	}
	return &r.Location, nil, nil
}

// Write exported values in context to files o a selected target folder
func (r *WindowsRequest) manageResults(stackResult auto.UpResult) error {
	return output.Write(stackResult, qenvsContext.GetResultsOutputPath(), map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputAdminUsername):     "adminusername",
		fmt.Sprintf("%s-%s", r.Prefix, outputAdminUserPassword): "adminuserpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):          "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword):      "userpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey):    "id_rsa",
		fmt.Sprintf("%s-%s", r.Prefix, outputHost):              "host",
	})
}

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func (r *WindowsRequest) createVirtualMachine(ctx *pulumi.Context,
	rg *resources.ResourceGroup, ni *network.NetworkInterface,
	location string, spotPrice *float64) (*compute.VirtualMachine, error) {
	if err := r.validateProfiles(); err != nil {
		return nil, err
	}
	adminPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pswd-adminuser"))
	if err != nil {
		return nil, err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputAdminUsername), pulumi.String(r.AdminUsername))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputAdminUserPassword), adminPasswd.Result)
	vmArgs := &compute.VirtualMachineArgs{
		VmName:            pulumi.String(qenvsContext.RunID()),
		Location:          pulumi.String(location),
		ResourceGroupName: rg.Name,
		NetworkProfile: compute.NetworkProfileArgs{
			NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
				compute.NetworkInterfaceReferenceArgs{
					Id: ni.ID(),
				},
			},
		},
		HardwareProfile: compute.HardwareProfileArgs{
			VmSize: pulumi.String(r.VMSize),
		},
		StorageProfile: compute.StorageProfileArgs{
			ImageReference: compute.ImageReferenceArgs{
				Publisher: pulumi.String("MicrosoftWindowsDesktop"),
				Offer:     pulumi.String(fmt.Sprintf("windows-%s", r.Version)),
				Sku:       pulumi.String(fmt.Sprintf("win%s-%s", r.Version, r.Feature)),
				Version:   pulumi.String("latest"),
			},
			OsDisk: compute.OSDiskArgs{
				Name:         pulumi.String(qenvsContext.RunID()),
				CreateOption: pulumi.String("FromImage"),
				Caching:      compute.CachingTypesReadWrite,
				ManagedDisk: compute.ManagedDiskParametersArgs{
					StorageAccountType: pulumi.String("Standard_LRS"),
				},
			},
		},
		OsProfile: compute.OSProfileArgs{
			AdminUsername: pulumi.String(r.AdminUsername),
			AdminPassword: adminPasswd.Result,
			ComputerName:  pulumi.String(qenvsContext.RunID()),
		},
		Tags: qenvsContext.ResourceTags(),
	}
	if spotPrice != nil {
		vmArgs.Priority = pulumi.String(prioritySpot)
		vmArgs.BillingProfile = compute.BillingProfileArgs{
			MaxPrice: pulumi.Float64(*spotPrice),
		}
	}

	return compute.NewVirtualMachine(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "vm"),
		vmArgs)
}

// Create networking resource required for spin the VM
func (r *WindowsRequest) createNetworking(ctx *pulumi.Context,
	rg *resources.ResourceGroup, location string) (*network.NetworkInterface,
	*network.PublicIPAddress, error) {
	vn, err := network.NewVirtualNetwork(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "vn"),
		&network.VirtualNetworkArgs{
			VirtualNetworkName: pulumi.String(qenvsContext.RunID()),
			AddressSpace: network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String(cidrVN),
				},
			},
			ResourceGroupName: rg.Name,
			Location:          pulumi.String(location),
			Tags:              qenvsContext.ResourceTags(),
		})
	if err != nil {
		return nil, nil, err
	}
	sn, err := network.NewSubnet(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "sn"),
		&network.SubnetArgs{
			SubnetName:         pulumi.String(qenvsContext.RunID()),
			ResourceGroupName:  rg.Name,
			VirtualNetworkName: vn.Name,
			AddressPrefixes: pulumi.StringArray{
				pulumi.String(cidrSN),
			},
		})
	if err != nil {
		return nil, nil, err
	}
	publicIP, err := network.NewPublicIPAddress(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pip"),
		&network.PublicIPAddressArgs{
			Location:                 pulumi.String(location),
			PublicIpAddressName:      pulumi.String(qenvsContext.RunID()),
			PublicIPAllocationMethod: pulumi.String("Static"),
			ResourceGroupName:        rg.Name,
			Tags:                     qenvsContext.ResourceTags(),
			// DnsSettings: network.PublicIPAddressDnsSettingsArgs{
			// 	DomainNameLabel: pulumi.String("qenvs"),
			// },
		})
	if err != nil {
		return nil, nil, err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputHost), publicIP.IpAddress)
	ni, err := network.NewNetworkInterface(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "ni"),
		&network.NetworkInterfaceArgs{
			NetworkInterfaceName: pulumi.String(qenvsContext.RunID()),
			Location:             pulumi.String(location),
			ResourceGroupName:    rg.Name,
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name:                      pulumi.String(qenvsContext.RunID()),
					PrivateIPAllocationMethod: pulumi.String("Dynamic"),
					PublicIPAddress: network.PublicIPAddressTypeArgs{
						Id: publicIP.ID(),
					},
					Subnet: network.SubnetTypeArgs{
						Id: sn.ID(),
					},
				},
			},
			Tags: qenvsContext.ResourceTags(),
		})
	if err != nil {
		return nil, nil, err
	}
	return ni, publicIP, nil
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
				"powershell -ExecutionPolicy Unrestricted -File %s %s -userPass \"%s\" -user %s -hostname %s -authorizedKey \"%s\"",
				scriptName,
				r.profilesAsParams(),
				password,
				r.Username,
				*hostname,
				authorizedKey,
			)
		}).(pulumi.StringOutput)
	// the post script will be executed as a extension
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
			Tags: qenvsContext.ResourceTags(),
		})
	return privateKey, vme, err
}

// Upload scrip to blob container to be used within Microsoft Compute extension
func (r *WindowsRequest) uploadScript(ctx *pulumi.Context,
	rg *resources.ResourceGroup, location string) (*storage.Blob, error) {
	sa, err := storage.NewStorageAccount(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "sa"),
		&storage.StorageAccountArgs{
			AccountName:           pulumi.String(qenvsContext.RunID()),
			Kind:                  pulumi.String("BlockBlobStorage"),
			ResourceGroupName:     rg.Name,
			Location:              pulumi.String(location),
			AllowBlobPublicAccess: pulumi.BoolPtr(true),
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Premium_LRS"),
			},
			Tags: qenvsContext.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	c, err := storage.NewBlobContainer(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "co"),
		&storage.BlobContainerArgs{
			ContainerName:     pulumi.String(qenvsContext.RunID()),
			AccountName:       sa.Name,
			ResourceGroupName: rg.Name,
			PublicAccess:      storage.PublicAccessBlob,
		})
	if err != nil {
		return nil, err
	}
	return storage.NewBlob(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "bl"),
		&storage.BlobArgs{
			AccountName:       sa.Name,
			ContainerName:     c.Name,
			ResourceGroupName: rg.Name,
			Source:            pulumi.NewStringAsset(string(RHQPCISetupScript)),
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
