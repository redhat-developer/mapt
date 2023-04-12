package windows

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/azure"
	azurePlugin "github.com/adrianriobo/qenvs/pkg/provider/azure/plugin"
	"github.com/adrianriobo/qenvs/pkg/provider/util/command"
	"github.com/adrianriobo/qenvs/pkg/provider/util/output"
	"github.com/adrianriobo/qenvs/pkg/provider/util/security"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/compute"
	"github.com/pulumi/pulumi-azure-native-sdk/network"
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/storage"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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
}

type syncRequest struct {
	name       string
	prefix     string
	username   string
	privatekey *string
}

func Create(r *WindowsRequest) (err error) {
	logging.Debug("Creating Windows Desktop")
	cs := manager.Stack{
		StackName:           qenvsContext.GetStackInstanceName(stackCreateWindowsDesktop),
		ProjectName:         qenvsContext.GetInstanceName(),
		BackedURL:           qenvsContext.GetBackedURL(),
		CloudProviderPlugin: azurePlugin.DefaultPlugin,
		DeployFunc:          r.deployer,
	}
	csResult, err := manager.UpStack(cs)
	if err != nil {
		return err
	}
	privatekey, err := r.manageResults(csResult, qenvsContext.GetResultsOutput())
	if err != nil {
		return err
	}
	logging.Debug("Windows Desktop has been created")
	logging.Debug("Sync Windows Desktop")
	sr := syncRequest{
		name:       qenvsContext.GetID(),
		prefix:     r.Prefix,
		username:   r.Username,
		privatekey: privatekey,
	}
	ss := manager.Stack{
		StackName:           qenvsContext.GetStackInstanceName(stackSyncWindowsDesktop),
		ProjectName:         qenvsContext.GetInstanceName(),
		BackedURL:           qenvsContext.GetBackedURL(),
		CloudProviderPlugin: azurePlugin.DefaultPlugin,
		DeployFunc:          sr.sync,
	}
	_, err = manager.UpStack(ss)
	if err != nil {
		logging.Debug("Windows Desktop is able to process workloads from now on")
	}
	return err
}

func Destroy() error {
	if err := azure.Destroy(
		qenvsContext.GetInstanceName(),
		qenvsContext.GetBackedURL(),
		qenvsContext.GetStackInstanceName(stackCreateWindowsDesktop)); err != nil {
		return err
	}
	return azure.Destroy(
		qenvsContext.GetInstanceName(),
		qenvsContext.GetBackedURL(),
		qenvsContext.GetStackInstanceName(stackSyncWindowsDesktop))
}

// Main function to deploy all requried resources to azure
func (r *WindowsRequest) deployer(ctx *pulumi.Context) error {
	rg, err := resources.NewResourceGroup(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "rg"),
		&resources.ResourceGroupArgs{
			Location:          pulumi.String(r.Location),
			ResourceGroupName: pulumi.String(qenvsContext.GetID()),
		})
	if err != nil {
		return err
	}
	ni, err := r.createNetworking(ctx, rg)
	if err != nil {
		return err
	}
	vm, err := r.createVirtualMachine(ctx, rg, ni)
	if err != nil {
		return err
	}
	return r.postInitSetup(ctx, rg, vm)
}

// This function works as a syncer for the created VM
// due to https://github.com/pulumi/pulumi-azure-native/issues/2336 we need to
// lookup for the ip after stack is created, also we need that value
// to run a remote command to wait for the instance to ensure it is healthy
// also this will export the host ip for the instance
func (r *syncRequest) sync(ctx *pulumi.Context) error {
	ip, err := network.LookupPublicIPAddress(ctx,
		&network.LookupPublicIPAddressArgs{
			PublicIpAddressName: r.name,
			ResourceGroupName:   r.name,
		})
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(qenvsContext.GetResultsOutput(), "host"), []byte(*ip.IpAddress), 0600)
	if err != nil {
		return err
	}
	_, err = remote.NewCommand(ctx,
		resourcesUtil.GetResourceName(r.prefix, azureWindowsDesktopID, "cmd"),
		&remote.CommandArgs{
			Connection: remote.ConnectionArgs{
				Host:           pulumi.String(*ip.IpAddress),
				PrivateKey:     pulumi.String(*r.privatekey),
				User:           pulumi.String(r.username),
				DialErrorLimit: pulumi.Int(-1),
			},
			Create: pulumi.String(command.CommandPing),
			Update: pulumi.String(command.CommandPing),
		})
	return err
}

// Write exported values in context to files o a selected target folder
func (r *WindowsRequest) manageResults(stackResult auto.UpResult,
	destinationFolder string) (*string, error) {
	if err := output.Write(stackResult, destinationFolder, map[string]string{
		fmt.Sprintf("%s-%s", r.Prefix, outputAdminUsername):     "adminusername",
		fmt.Sprintf("%s-%s", r.Prefix, outputAdminUserPassword): "adminuserpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUsername):          "username",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPassword):      "userpassword",
		fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey):    "id_rsa",
	}); err != nil {
		return nil, err
	}
	privatekey, ok := stackResult.Outputs[fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey)].Value.(string)
	if !ok {
		return nil, fmt.Errorf("error getting private key")
	}
	return &privatekey, nil
}

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func (r *WindowsRequest) createVirtualMachine(ctx *pulumi.Context,
	rg *resources.ResourceGroup, ni *network.NetworkInterface) (*compute.VirtualMachine, error) {
	adminPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pswd-adminuser"))
	if err != nil {
		return nil, err

	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputAdminUsername), pulumi.String(r.AdminUsername))
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputAdminUserPassword), adminPasswd.Result)
	return compute.NewVirtualMachine(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "vm"),
		&compute.VirtualMachineArgs{
			VmName:            pulumi.String(qenvsContext.GetID()),
			Location:          rg.Location,
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
					Name:         pulumi.String(qenvsContext.GetID()),
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
				ComputerName:  pulumi.String(qenvsContext.GetID()),
			},
			Tags: qenvsContext.GetTags(),
		})
}

// Create networking resource required for spin the VM
func (r *WindowsRequest) createNetworking(ctx *pulumi.Context,
	rg *resources.ResourceGroup) (*network.NetworkInterface, error) {
	vn, err := network.NewVirtualNetwork(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "vn"),
		&network.VirtualNetworkArgs{
			VirtualNetworkName: pulumi.String(qenvsContext.GetID()),
			AddressSpace: network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String(cidrVN),
				},
			},
			ResourceGroupName: rg.Name,
			Location:          rg.Location,
			Tags:              qenvsContext.GetTags(),
		})
	if err != nil {
		return nil, err
	}
	sn, err := network.NewSubnet(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "sn"),
		&network.SubnetArgs{
			SubnetName:         pulumi.String(qenvsContext.GetID()),
			ResourceGroupName:  rg.Name,
			VirtualNetworkName: vn.Name,
			AddressPrefixes: pulumi.StringArray{
				pulumi.String(cidrSN),
			},
		})
	if err != nil {
		return nil, err
	}
	publicIP, err := network.NewPublicIPAddress(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pip"),
		&network.PublicIPAddressArgs{
			Location:            rg.Location,
			PublicIpAddressName: pulumi.String(qenvsContext.GetID()),
			ResourceGroupName:   rg.Name,
			Tags:                qenvsContext.GetTags(),
			// DnsSettings: network.PublicIPAddressDnsSettingsArgs{
			// 	DomainNameLabel: pulumi.String("qenvs"),
			// },
		})
	if err != nil {
		return nil, err
	}
	return network.NewNetworkInterface(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "ni"),
		&network.NetworkInterfaceArgs{
			NetworkInterfaceName: pulumi.String(qenvsContext.GetID()),
			Location:             rg.Location,
			ResourceGroupName:    rg.Name,
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name:                      pulumi.String(qenvsContext.GetID()),
					PrivateIPAllocationMethod: pulumi.String("Dynamic"),
					PublicIPAddress: network.PublicIPAddressTypeArgs{
						Id: publicIP.ID(),
					},
					Subnet: network.SubnetTypeArgs{
						Id: sn.ID(),
					},
				},
			},
			Tags: qenvsContext.GetTags(),
		})
}

// run a post script to setup the machine as expected according to rhqp-ci-setup.ps1
// it also exports to pulumi context user name, user password and user privatekey
func (r *WindowsRequest) postInitSetup(ctx *pulumi.Context, rg *resources.ResourceGroup,
	vm *compute.VirtualMachine) error {
	userPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "pswd-user"))
	if err != nil {
		return err

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
		return err
	}
	ctx.Export(fmt.Sprintf("%s-%s", r.Prefix, outputUserPrivateKey), privateKey.PrivateKeyPem)
	// upload the script to a ephemeral blob container
	b, err := r.uploadScript(ctx, rg)
	if err != nil {
		return err
	}
	// the post script command will be generated based on generated data as parameters
	setupCommand := pulumi.All(userPasswd.Result, privateKey.PublicKeyOpenssh, vm.OsProfile.ComputerName()).ApplyT(
		func(args []interface{}) string {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			hostname := args[2].(*string)
			return fmt.Sprintf(
				"powershell -ExecutionPolicy Unrestricted -File %s -userPass \"%s\" -user %s -hostname %s -authorizedKey \"%s\"",
				scriptName,
				password,
				r.Username,
				*hostname,
				authorizedKey)
		}).(pulumi.StringOutput)
	// the post script will be executed as a extension
	_, err = compute.NewVirtualMachineExtension(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "ext"),
		&compute.VirtualMachineExtensionArgs{
			ResourceGroupName:  rg.Name,
			Location:           rg.Location,
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
			Tags: qenvsContext.GetTags(),
		})
	return err
}

// Upload scrip to blob container to be used within Microsoft Compute extension
func (r *WindowsRequest) uploadScript(ctx *pulumi.Context,
	rg *resources.ResourceGroup) (*storage.Blob, error) {
	sa, err := storage.NewStorageAccount(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "sa"),
		&storage.StorageAccountArgs{
			AccountName:       pulumi.String(qenvsContext.GetID()),
			Kind:              pulumi.String("BlockBlobStorage"),
			ResourceGroupName: rg.Name,
			Location:          rg.Location,
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Premium_LRS"),
			},
			Tags: qenvsContext.GetTags(),
		})
	if err != nil {
		return nil, err
	}
	c, err := storage.NewBlobContainer(ctx,
		resourcesUtil.GetResourceName(r.Prefix, azureWindowsDesktopID, "co"),
		&storage.BlobContainerArgs{
			ContainerName:     pulumi.String(qenvsContext.GetID()),
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
