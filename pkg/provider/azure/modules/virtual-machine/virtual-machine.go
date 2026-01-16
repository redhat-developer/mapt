package virtualmachine

import (
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi-azure-native-sdk/compute/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
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

type VirtualMachineArgs struct {
	Prefix          string
	ComponentID     string
	ResourceGroup   *resources.ResourceGroup
	NetworkInteface *network.NetworkInterface
	VMSize          string

	SpotPrice *float64
	// community galary image ID
	Image *data.ImageReference
	// Windows required
	AdminUsername string
	// Linux required
	PrivateKey  *tls.PrivateKey
	AdminPasswd *random.RandomPassword
	// Only required if we need to set userdata
	UserDataAsBase64 pulumi.StringPtrInput
	Location         string
}

type VirtualMachine = *compute.VirtualMachine

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *VirtualMachineArgs) (VirtualMachine, error) {
	ira, err := convertImageRef(mCtx, *args.Image, args.Location)
	if err != nil {
		return nil, err
	}
	vmArgs := &compute.VirtualMachineArgs{
		VmName:            pulumi.String(mCtx.RunID()),
		Location:          pulumi.String(args.Location),
		ResourceGroupName: args.ResourceGroup.Name,
		UserData:          args.UserDataAsBase64,
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
			ImageReference: ira,
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

		OsProfile: osProfile(mCtx.RunID(), args),
		Tags:      mCtx.ResourceTags(),
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

func convertImageRef(mCtx *mc.Context, i data.ImageReference, location string) (*compute.ImageReferenceArgs, error) {
	if len(i.CommunityImageID) > 0 {
		return &compute.ImageReferenceArgs{
			CommunityGalleryImageId: pulumi.String(i.CommunityImageID),
		}, nil
	}
	if len(i.SharedImageID) > 0 {
		if isSelfOwned(&i.SharedImageID) {
			return &compute.ImageReferenceArgs{
				Id: pulumi.String(i.SharedImageID),
			}, nil
		}
		return &compute.ImageReferenceArgs{
			SharedGalleryImageId: pulumi.String(i.SharedImageID),
		}, nil

	}
	finalSku, err := data.SkuG2Support(mCtx.Context(), location, i.Publisher, i.Offer, i.Sku)
	if err != nil {
		return nil, err
	}
	return &compute.ImageReferenceArgs{
		Publisher: pulumi.String(i.Publisher),
		Offer:     pulumi.String(i.Offer),
		Sku:       pulumi.String(finalSku),
		Version:   pulumi.String("latest"),
	}, nil
}

func isSelfOwned(sharedImageId *string) bool {
	sharedImageParams := strings.Split(*sharedImageId, "/")
	return os.Getenv("AZURE_SUBSCRIPTION_ID") == sharedImageParams[2]
}
