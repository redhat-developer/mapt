package virtualmachine

import (
	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/util/security"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const prioritySpot = "Spot"

type VirtualMachineRequest struct {
	Prefix          string
	ComponentID     string
	ResourceGroup   *resources.ResourceGroup
	NetworkInteface *network.NetworkInterface
	VMSize          string
	Publisher       string
	Offer           string
	Sku             string
	Version         string
	AdminUsername   string
	SpotPrice       *float64
}

type VirtualMachine struct {
	VM            *compute.VirtualMachine
	AdminPassword *random.RandomPassword
}

// Create virtual machine based on request + export to context
// adminusername and adminuserpassword
func (r *VirtualMachineRequest) Create(ctx *pulumi.Context) (*VirtualMachine, error) {
	adminPasswd, err := security.CreatePassword(
		ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "pswd-adminuser"))
	if err != nil {
		return nil, err
	}
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
			ComputerName:  pulumi.String(maptContext.RunID()),
		},
		Tags: maptContext.ResourceTags(),
	}
	if r.SpotPrice != nil {
		vmArgs.Priority = pulumi.String(prioritySpot)
		vmArgs.BillingProfile = compute.BillingProfileArgs{
			MaxPrice: pulumi.Float64(*r.SpotPrice),
		}
	}
	vm, err := compute.NewVirtualMachine(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "vm"),
		vmArgs)
	if err != nil {
		return nil, err
	}
	return &VirtualMachine{
		VM:            vm,
		AdminPassword: adminPasswd,
	}, nil
}
