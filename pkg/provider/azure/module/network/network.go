package network

import (
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type NetworkRequest struct {
	Prefix        string
	ComponentID   string
	ResourceGroup *resources.ResourceGroup
}

type Network struct {
	Network          *network.VirtualNetwork
	PublicSubnet     *network.Subnet
	NetworkInterface *network.NetworkInterface
	PublicIP         *network.PublicIPAddress
}

// Create networking resource required for spin the VM
func (r *NetworkRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*Network, error) {
	vn, err := network.NewVirtualNetwork(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "vn"),
		&network.VirtualNetworkArgs{
			VirtualNetworkName: pulumi.String(mCtx.RunID()),
			AddressSpace: network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String(cidrVN),
				},
			},
			ResourceGroupName: r.ResourceGroup.Name,
			Location:          r.ResourceGroup.Location,
			Tags:              mCtx.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	sn, err := network.NewSubnet(ctx,
		resourcesUtil.GetResourceName(r.Prefix, r.ComponentID, "sn"),
		&network.SubnetArgs{
			SubnetName:         pulumi.String(mCtx.RunID()),
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
			PublicIpAddressName:      pulumi.String(mCtx.RunID()),
			PublicIPAllocationMethod: pulumi.String("Static"),
			ResourceGroupName:        r.ResourceGroup.Name,
			Tags:                     mCtx.ResourceTags(),
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
			NetworkInterfaceName: pulumi.String(mCtx.RunID()),
			Location:             r.ResourceGroup.Location,
			ResourceGroupName:    r.ResourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name:                      pulumi.String(mCtx.RunID()),
					PrivateIPAllocationMethod: pulumi.String("Dynamic"),
					PublicIPAddress: network.PublicIPAddressTypeArgs{
						Id: publicIP.ID(),
					},
					Subnet: network.SubnetTypeArgs{
						Id: sn.ID(),
					},
				},
			},
			Tags: mCtx.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	return &Network{
		NetworkInterface: ni,
		PublicIP:         publicIP,
		Network:          vn,
		PublicSubnet:     sn,
	}, nil
}
