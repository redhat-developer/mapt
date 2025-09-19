package virtualmachine

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-azure-native-sdk/compute/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	prioritySpot     = "Spot"
	diskSize     int = 200
)

type VirtualMachineArgs struct {
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
	Userdata   pulumi.StringInput
	Location   string
	DiskSizeGB int
}

type VirtualMachine = *compute.VirtualMachine

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *VirtualMachineArgs) (VirtualMachine, error) {
	var imageReferenceArgs compute.ImageReferenceArgs
	if len(args.ImageID) > 0 {
		imageReferenceArgs = getImageRefArgs(args.ImageID)
	} else {
		finalSku, err := data.SkuG2Support(args.Location, args.Publisher, args.Offer, args.Sku)
		if err != nil {
			return nil, err
		}
		imageReferenceArgs = compute.ImageReferenceArgs{
			Publisher: pulumi.String(args.Publisher),
			Offer:     pulumi.String(args.Offer),
			Sku:       pulumi.String(finalSku),
			Version:   pulumi.String("latest"),
		}
	}
	vmArgs := &compute.VirtualMachineArgs{
		VmName:            pulumi.String(mCtx.RunID()),
		Location:          pulumi.String(args.Location),
		ResourceGroupName: args.ResourceGroup.Name,
		NetworkProfile: compute.NetworkProfileArgs{
			NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
				compute.NetworkInterfaceReferenceArgs{
					Id: args.NetworkInteface.ID(),
				},
			},
		},
		HardwareProfile: compute.HardwareProfileArgs{
			VmSize: pulumi.String(args.VMSize),
		},
		StorageProfile: compute.StorageProfileArgs{
			ImageReference: imageReferenceArgs,
			OsDisk: compute.OSDiskArgs{
				Name:         pulumi.String(mCtx.RunID()),
				DiskSizeGB:   util.If(args.DiskSizeGB > 0, pulumi.Int(args.DiskSizeGB), pulumi.Int(diskSize)),
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

		OsProfile: osProfile(mCtx.RunID(), args),
		Tags:      mCtx.ResourceTags(),
		UserData:  args.Userdata,
	}
	if args.SpotPrice != nil {
		vmArgs.Priority = pulumi.String(prioritySpot)
		vmArgs.BillingProfile = compute.BillingProfileArgs{
			MaxPrice: pulumi.Float64(*args.SpotPrice),
		}
	}
	logging.Debug("About to create the VM with compute.NewVirtualMachine")
	return compute.NewVirtualMachine(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "vm"),
		vmArgs)
}

func osProfile(computerName string, args *VirtualMachineArgs) compute.OSProfileArgs {
	osProfile := compute.OSProfileArgs{
		AdminUsername: pulumi.String(args.AdminUsername),
		ComputerName:  pulumi.String(computerName),
	}
	if args.AdminPasswd != nil {
		osProfile.AdminPassword = args.AdminPasswd.Result
	}
	if args.PrivateKey != nil {
		osProfile.LinuxConfiguration = &compute.LinuxConfigurationArgs{
			PatchSettings:                 compute.LinuxPatchSettingsArgs{},
			DisablePasswordAuthentication: pulumi.Bool(true),
			Ssh: &compute.SshConfigurationArgs{
				PublicKeys: compute.SshPublicKeyTypeArray{
					&compute.SshPublicKeyTypeArgs{
						KeyData: args.PrivateKey.PublicKeyOpenssh,
						Path:    pulumi.String(fmt.Sprintf("/home/%s/.ssh/authorized_keys", args.AdminUsername)),
					},
				},
			},
		}
	}
	return osProfile
}

func getImageRefArgs(imageID string) compute.ImageReferenceArgs {
	if strings.Contains(imageID, "Community") {
		return compute.ImageReferenceArgs{
			CommunityGalleryImageId: pulumi.String(imageID),
		}
	}
	return compute.ImageReferenceArgs{
		Id: pulumi.String(imageID),
	}
}
