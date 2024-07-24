package network

import (
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type NetworkRequest struct {
	Prefix        string
	ComponentID   string
	ResourceGroup *resources.ResourceGroup
}

type Netowork struct {
	NetworkInterface *network.NetworkInterface
	PublicIP         *network.PublicIPAddress
}

// Create networking resource required for spin the VM
func (r *NetworkRequest) Create(ctx *pulumi.Context) (*Netowork, error) {
	vn, err := network.NewVirtualNetwork(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "vn"),
		&network.VirtualNetworkArgs{
			VirtualNetworkName: pulumi.String(maptContext.RunID()),
			AddressSpace: network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String(cidrVN),
				},
			},
			ResourceGroupName: r.ResourceGroup.Name,
			Location:          r.ResourceGroup.Location,
			Tags:              maptContext.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	sn, err := network.NewSubnet(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "sn"),
		&network.SubnetArgs{
			SubnetName:         pulumi.String(maptContext.RunID()),
			ResourceGroupName:  r.ResourceGroup.Name,
			VirtualNetworkName: vn.Name,
			AddressPrefixes: pulumi.StringArray{
				pulumi.String(cidrSN),
			},
		})
	if err != nil {
		return nil, err
	}
	publicIP, err := network.NewPublicIPAddress(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "pip"),
		&network.PublicIPAddressArgs{
			Location:                 r.ResourceGroup.Location,
			PublicIpAddressName:      pulumi.String(maptContext.RunID()),
			PublicIPAllocationMethod: pulumi.String("Static"),
			ResourceGroupName:        r.ResourceGroup.Name,
			Tags:                     maptContext.ResourceTags(),
			// DnsSettings: network.PublicIPAddressDnsSettingsArgs{
			// 	DomainNameLabel: pulumi.String("mapt"),
			// },
		})
	if err != nil {
		return nil, err
	}
	ni, err := network.NewNetworkInterface(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "ni"),
		&network.NetworkInterfaceArgs{
			NetworkInterfaceName: pulumi.String(maptContext.RunID()),
			Location:             r.ResourceGroup.Location,
			ResourceGroupName:    r.ResourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name:                      pulumi.String(maptContext.RunID()),
					PrivateIPAllocationMethod: pulumi.String("Dynamic"),
					PublicIPAddress: network.PublicIPAddressTypeArgs{
						Id: publicIP.ID(),
					},
					Subnet: network.SubnetTypeArgs{
						Id: sn.ID(),
					},
				},
			},
			Tags: maptContext.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	return &Netowork{
		NetworkInterface: ni,
		PublicIP:         publicIP,
	}, nil
}
