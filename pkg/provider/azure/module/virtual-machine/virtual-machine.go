package virtualmachine

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
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
	// Windows required
	AdminUsername string
	// Linux required
	PrivateKey  *tls.PrivateKey
	AdminPasswd *random.RandomPassword
}

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func (r *VirtualMachineRequest) Create(ctx *pulumi.Context) (*compute.VirtualMachine, error) {
	vmArgs := &compute.VirtualMachineArgs{
		VmName:            pulumi.String(maptContext.RunID()),
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
			ImageReference: compute.ImageReferenceArgs{
				Publisher: pulumi.String(r.Publisher),
				Offer:     pulumi.String(r.Offer),
				Sku:       pulumi.String(r.Sku),
				Version:   pulumi.String("latest"),
			},
			OsDisk: compute.OSDiskArgs{
				Name:         pulumi.String(maptContext.RunID()),
				DiskSizeGB:   pulumi.Int(diskSize),
				CreateOption: pulumi.String("FromImage"),
				Caching:      compute.CachingTypesReadWrite,
				ManagedDisk: compute.ManagedDiskParametersArgs{
					StorageAccountType: pulumi.String("Standard_LRS"),
				},
			},
		},
		OsProfile: r.osProfile(),

		Tags: maptContext.ResourceTags(),
	}
	if r.SpotPrice != nil {
		vmArgs.Priority = pulumi.String(prioritySpot)
		vmArgs.BillingProfile = compute.BillingProfileArgs{
			MaxPrice: pulumi.Float64(*r.SpotPrice),
		}
	}
	return compute.NewVirtualMachine(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "vm"),
		vmArgs)
}

func (r *VirtualMachineRequest) osProfile() compute.OSProfileArgs {
	osProfile := compute.OSProfileArgs{
		AdminUsername: pulumi.String(r.AdminUsername),
		ComputerName:  pulumi.String(maptContext.RunID()),
	}
	if r.AdminPasswd != nil {
		osProfile.AdminPassword = r.AdminPasswd.Result
	}
	if r.PrivateKey != nil {
		osProfile.LinuxConfiguration = &compute.LinuxConfigurationArgs{
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
