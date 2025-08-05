package network

import (
	"github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	securityGroup "github.com/redhat-developer/mapt/pkg/provider/azure/services/network/security-group"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type NetworkArgs struct {
	Prefix        string
	ComponentID   string
	ResourceGroup *resources.ResourceGroup
	Location      *string
	SecurityGroup securityGroup.SecurityGroup
}

type Network struct {
	Network          *network.VirtualNetwork
	PublicSubnet     *network.Subnet
	NetworkInterface *network.NetworkInterface
	PublicIP         *network.PublicIPAddress
}

// Create networking resource required for spin the VM
func Create(ctx *pulumi.Context, mCtx *mc.Context, args *NetworkArgs) (*Network, error) {
	vn, err := network.NewVirtualNetwork(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "vn"),
		&network.VirtualNetworkArgs{
			VirtualNetworkName: pulumi.String(mCtx.RunID()),
			AddressSpace: network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String(cidrVN),
				},
			},
			ResourceGroupName: args.ResourceGroup.Name,
			Location:          pulumi.String(*args.Location),
			Tags:              mCtx.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	sn, err := network.NewSubnet(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "sn"),
		&network.SubnetArgs{
			SubnetName:         pulumi.String(mCtx.RunID()),
			ResourceGroupName:  args.ResourceGroup.Name,
			VirtualNetworkName: vn.Name,
			AddressPrefixes: pulumi.StringArray{
				pulumi.String(cidrSN),
			},
		})
	if err != nil {
		return nil, err
	}
	publicIP, err := network.NewPublicIPAddress(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "pip"),
		&network.PublicIPAddressArgs{
			Location:                 pulumi.String(*args.Location),
			PublicIpAddressName:      pulumi.String(mCtx.RunID()),
			PublicIPAllocationMethod: pulumi.String("Static"),
			Sku: &network.PublicIPAddressSkuArgs{
				Name: pulumi.String("Standard"),
			},
			ResourceGroupName: args.ResourceGroup.Name,
			Tags:              mCtx.ResourceTags(),
			// DnsSettings: network.PublicIPAddressDnsSettingsArgs{
			// 	DomainNameLabel: pulumi.String("mapt"),
			// },
		})
	if err != nil {
		return nil, err
	}
	ni, err := network.NewNetworkInterface(ctx,
		resourcesUtil.GetResourceName(args.Prefix, args.ComponentID, "ni"),
		&network.NetworkInterfaceArgs{
			NetworkInterfaceName: pulumi.String(mCtx.RunID()),
			Location:             pulumi.String(*args.Location),
			ResourceGroupName:    args.ResourceGroup.Name,
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
			NetworkSecurityGroup: &network.NetworkSecurityGroupTypeArgs{
				Id: args.SecurityGroup.ID(),
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
