package virtualmachine

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	prioritySpot     = "Spot"
	diskSize     int = 200
)

type VirtualMachineRequest struct {
	Prefix          string
	ComponentID     string
	ResourceGroup   *resources.ResourceGroup
	NetworkInteface *network.NetworkInterface
	VMSize          string
	Publisher       string
	Offer           string
	Sku             string
	SpotPrice       *float64
	// community galary image ID
	ImageID string
	// Windows required
	AdminUsername string
	// Linux required
	PrivateKey  *tls.PrivateKey
	AdminPasswd *random.RandomPassword
	// Linux optional
	Userdata string
	Location string
}

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func (r *VirtualMachineRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*compute.VirtualMachine, error) {
	var imageReferenceArgs compute.ImageReferenceArgs
	if len(r.ImageID) > 0 {
		imageReferenceArgs = compute.ImageReferenceArgs{
			CommunityGalleryImageId: pulumi.String(r.ImageID)}
	} else {
		finalSku, err := data.SkuG2Support(r.Location, r.Publisher, r.Offer, r.Sku)
		if err != nil {
			return nil, err
		}
		imageReferenceArgs = compute.ImageReferenceArgs{
			Publisher: pulumi.String(r.Publisher),
			Offer:     pulumi.String(r.Offer),
			Sku:       pulumi.String(finalSku),
			Version:   pulumi.String("latest"),
		}
	}
	vmArgs := &compute.VirtualMachineArgs{
		VmName:            pulumi.String(mCtx.RunID()),
		Location:          r.ResourceGroup.Location,
		ResourceGroupName: r.ResourceGroup.Name,
		NetworkProfile: compute.NetworkProfileArgs{
			NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
				compute.NetworkInterfaceReferenceArgs{
					Id: r.NetworkInteface.ID(),
				},
			},
		},
		HardwareProfile: compute.HardwareProfileArgs{
			VmSize: pulumi.String(r.VMSize),
		},
		StorageProfile: compute.StorageProfileArgs{
			ImageReference: imageReferenceArgs,
			OsDisk: compute.OSDiskArgs{
				Name:         pulumi.String(mCtx.RunID()),
				DiskSizeGB:   pulumi.Int(diskSize),
				CreateOption: pulumi.String("FromImage"),
				Caching:      compute.CachingTypesReadWrite,
				ManagedDisk: compute.ManagedDiskParametersArgs{
					StorageAccountType: pulumi.String("Standard_LRS"),
				},
			},
		},
		// Try to improve provisioning time
		DiagnosticsProfile: compute.DiagnosticsProfileArgs{
			BootDiagnostics: compute.BootDiagnosticsArgs{
				Enabled: pulumi.Bool(false),
			},
		},

		OsProfile: r.osProfile(mCtx.RunID()),
		Tags:      mCtx.ResourceTags(),
	}
	if r.SpotPrice != nil {
		vmArgs.Priority = pulumi.String(prioritySpot)
		vmArgs.BillingProfile = compute.BillingProfileArgs{
			MaxPrice: pulumi.Float64(*r.SpotPrice),
		}
	}
	if len(r.Userdata) > 0 {
		vmArgs.UserData = pulumi.String(r.Userdata)
	}
	logging.Debug("About to create the VM with compute.NewVirtualMachine")
	return compute.NewVirtualMachine(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "vm"),
		vmArgs)
}

func (r *VirtualMachineRequest) osProfile(computerName string) compute.OSProfileArgs {
	osProfile := compute.OSProfileArgs{
		AdminUsername: pulumi.String(r.AdminUsername),
		ComputerName:  pulumi.String(computerName),
	}
	if r.AdminPasswd != nil {
		osProfile.AdminPassword = r.AdminPasswd.Result
	}
	if r.PrivateKey != nil {
		osProfile.LinuxConfiguration = &compute.LinuxConfigurationArgs{
			PatchSettings:                 compute.LinuxPatchSettingsArgs{},
			DisablePasswordAuthentication: pulumi.Bool(true),
			Ssh: &compute.SshConfigurationArgs{
				PublicKeys: compute.SshPublicKeyTypeArray{
					&compute.SshPublicKeyTypeArgs{
						KeyData: r.PrivateKey.PublicKeyOpenssh,
						Path:    pulumi.String(fmt.Sprintf("/home/%s/.ssh/authorized_keys", r.AdminUsername)),
					},
				},
			},
		}
	}
	return osProfile
}
