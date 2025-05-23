// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package network

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Peerings in a virtual network resource.
//
// Uses Azure REST API version 2023-02-01. In version 1.x of the Azure Native provider, it used API version 2020-11-01.
//
// Other available API versions: 2019-06-01, 2023-04-01, 2023-05-01, 2023-06-01, 2023-09-01, 2023-11-01, 2024-01-01, 2024-03-01, 2024-05-01.
type VirtualNetworkPeering struct {
	pulumi.CustomResourceState

	// Whether the forwarded traffic from the VMs in the local virtual network will be allowed/disallowed in remote virtual network.
	AllowForwardedTraffic pulumi.BoolPtrOutput `pulumi:"allowForwardedTraffic"`
	// If gateway links can be used in remote virtual networking to link to this virtual network.
	AllowGatewayTransit pulumi.BoolPtrOutput `pulumi:"allowGatewayTransit"`
	// Whether the VMs in the local virtual network space would be able to access the VMs in remote virtual network space.
	AllowVirtualNetworkAccess pulumi.BoolPtrOutput `pulumi:"allowVirtualNetworkAccess"`
	// If we need to verify the provisioning state of the remote gateway.
	DoNotVerifyRemoteGateways pulumi.BoolPtrOutput `pulumi:"doNotVerifyRemoteGateways"`
	// A unique read-only string that changes whenever the resource is updated.
	Etag pulumi.StringOutput `pulumi:"etag"`
	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name pulumi.StringPtrOutput `pulumi:"name"`
	// The status of the virtual network peering.
	PeeringState pulumi.StringPtrOutput `pulumi:"peeringState"`
	// The peering sync status of the virtual network peering.
	PeeringSyncLevel pulumi.StringPtrOutput `pulumi:"peeringSyncLevel"`
	// The provisioning state of the virtual network peering resource.
	ProvisioningState pulumi.StringOutput `pulumi:"provisioningState"`
	// The reference to the address space peered with the remote virtual network.
	RemoteAddressSpace AddressSpaceResponsePtrOutput `pulumi:"remoteAddressSpace"`
	// The reference to the remote virtual network's Bgp Communities.
	RemoteBgpCommunities VirtualNetworkBgpCommunitiesResponsePtrOutput `pulumi:"remoteBgpCommunities"`
	// The reference to the remote virtual network. The remote virtual network can be in the same or different region (preview). See here to register for the preview and learn more (https://docs.microsoft.com/en-us/azure/virtual-network/virtual-network-create-peering).
	RemoteVirtualNetwork SubResourceResponsePtrOutput `pulumi:"remoteVirtualNetwork"`
	// The reference to the current address space of the remote virtual network.
	RemoteVirtualNetworkAddressSpace AddressSpaceResponsePtrOutput `pulumi:"remoteVirtualNetworkAddressSpace"`
	// The reference to the remote virtual network's encryption
	RemoteVirtualNetworkEncryption VirtualNetworkEncryptionResponseOutput `pulumi:"remoteVirtualNetworkEncryption"`
	// The resourceGuid property of the Virtual Network peering resource.
	ResourceGuid pulumi.StringOutput `pulumi:"resourceGuid"`
	// Resource type.
	Type pulumi.StringPtrOutput `pulumi:"type"`
	// If remote gateways can be used on this virtual network. If the flag is set to true, and allowGatewayTransit on remote peering is also true, virtual network will use gateways of remote virtual network for transit. Only one peering can have this flag set to true. This flag cannot be set if virtual network already has a gateway.
	UseRemoteGateways pulumi.BoolPtrOutput `pulumi:"useRemoteGateways"`
}

// NewVirtualNetworkPeering registers a new resource with the given unique name, arguments, and options.
func NewVirtualNetworkPeering(ctx *pulumi.Context,
	name string, args *VirtualNetworkPeeringArgs, opts ...pulumi.ResourceOption) (*VirtualNetworkPeering, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.ResourceGroupName == nil {
		return nil, errors.New("invalid value for required argument 'ResourceGroupName'")
	}
	if args.VirtualNetworkName == nil {
		return nil, errors.New("invalid value for required argument 'VirtualNetworkName'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("azure-native:network/v20160601:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20160901:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20161201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20170301:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20170601:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20170801:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20170901:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20171001:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20171101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20180101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20180201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20180401:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20180601:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20180701:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20180801:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20181001:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20181101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20181201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20190201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20190401:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20190601:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20190701:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20190801:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20190901:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20191101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20191201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20200301:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20200401:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20200501:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20200601:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20200701:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20200801:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20201101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20210201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20210301:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20210501:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20210801:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220501:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220701:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220901:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20221101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230201:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230401:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230501:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230601:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230901:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20231101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240101:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240301:VirtualNetworkPeering"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240501:VirtualNetworkPeering"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource VirtualNetworkPeering
	err := ctx.RegisterResource("azure-native:network:VirtualNetworkPeering", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetVirtualNetworkPeering gets an existing VirtualNetworkPeering resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetVirtualNetworkPeering(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *VirtualNetworkPeeringState, opts ...pulumi.ResourceOption) (*VirtualNetworkPeering, error) {
	var resource VirtualNetworkPeering
	err := ctx.ReadResource("azure-native:network:VirtualNetworkPeering", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering VirtualNetworkPeering resources.
type virtualNetworkPeeringState struct {
}

type VirtualNetworkPeeringState struct {
}

func (VirtualNetworkPeeringState) ElementType() reflect.Type {
	return reflect.TypeOf((*virtualNetworkPeeringState)(nil)).Elem()
}

type virtualNetworkPeeringArgs struct {
	// Whether the forwarded traffic from the VMs in the local virtual network will be allowed/disallowed in remote virtual network.
	AllowForwardedTraffic *bool `pulumi:"allowForwardedTraffic"`
	// If gateway links can be used in remote virtual networking to link to this virtual network.
	AllowGatewayTransit *bool `pulumi:"allowGatewayTransit"`
	// Whether the VMs in the local virtual network space would be able to access the VMs in remote virtual network space.
	AllowVirtualNetworkAccess *bool `pulumi:"allowVirtualNetworkAccess"`
	// If we need to verify the provisioning state of the remote gateway.
	DoNotVerifyRemoteGateways *bool `pulumi:"doNotVerifyRemoteGateways"`
	// Resource ID.
	Id *string `pulumi:"id"`
	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `pulumi:"name"`
	// The status of the virtual network peering.
	PeeringState *string `pulumi:"peeringState"`
	// The peering sync status of the virtual network peering.
	PeeringSyncLevel *string `pulumi:"peeringSyncLevel"`
	// The reference to the address space peered with the remote virtual network.
	RemoteAddressSpace *AddressSpace `pulumi:"remoteAddressSpace"`
	// The reference to the remote virtual network's Bgp Communities.
	RemoteBgpCommunities *VirtualNetworkBgpCommunities `pulumi:"remoteBgpCommunities"`
	// The reference to the remote virtual network. The remote virtual network can be in the same or different region (preview). See here to register for the preview and learn more (https://docs.microsoft.com/en-us/azure/virtual-network/virtual-network-create-peering).
	RemoteVirtualNetwork *SubResource `pulumi:"remoteVirtualNetwork"`
	// The reference to the current address space of the remote virtual network.
	RemoteVirtualNetworkAddressSpace *AddressSpace `pulumi:"remoteVirtualNetworkAddressSpace"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// Parameter indicates the intention to sync the peering with the current address space on the remote vNet after it's updated.
	SyncRemoteAddressSpace *string `pulumi:"syncRemoteAddressSpace"`
	// Resource type.
	Type *string `pulumi:"type"`
	// If remote gateways can be used on this virtual network. If the flag is set to true, and allowGatewayTransit on remote peering is also true, virtual network will use gateways of remote virtual network for transit. Only one peering can have this flag set to true. This flag cannot be set if virtual network already has a gateway.
	UseRemoteGateways *bool `pulumi:"useRemoteGateways"`
	// The name of the virtual network.
	VirtualNetworkName string `pulumi:"virtualNetworkName"`
	// The name of the peering.
	VirtualNetworkPeeringName *string `pulumi:"virtualNetworkPeeringName"`
}

// The set of arguments for constructing a VirtualNetworkPeering resource.
type VirtualNetworkPeeringArgs struct {
	// Whether the forwarded traffic from the VMs in the local virtual network will be allowed/disallowed in remote virtual network.
	AllowForwardedTraffic pulumi.BoolPtrInput
	// If gateway links can be used in remote virtual networking to link to this virtual network.
	AllowGatewayTransit pulumi.BoolPtrInput
	// Whether the VMs in the local virtual network space would be able to access the VMs in remote virtual network space.
	AllowVirtualNetworkAccess pulumi.BoolPtrInput
	// If we need to verify the provisioning state of the remote gateway.
	DoNotVerifyRemoteGateways pulumi.BoolPtrInput
	// Resource ID.
	Id pulumi.StringPtrInput
	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name pulumi.StringPtrInput
	// The status of the virtual network peering.
	PeeringState pulumi.StringPtrInput
	// The peering sync status of the virtual network peering.
	PeeringSyncLevel pulumi.StringPtrInput
	// The reference to the address space peered with the remote virtual network.
	RemoteAddressSpace AddressSpacePtrInput
	// The reference to the remote virtual network's Bgp Communities.
	RemoteBgpCommunities VirtualNetworkBgpCommunitiesPtrInput
	// The reference to the remote virtual network. The remote virtual network can be in the same or different region (preview). See here to register for the preview and learn more (https://docs.microsoft.com/en-us/azure/virtual-network/virtual-network-create-peering).
	RemoteVirtualNetwork SubResourcePtrInput
	// The reference to the current address space of the remote virtual network.
	RemoteVirtualNetworkAddressSpace AddressSpacePtrInput
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput
	// Parameter indicates the intention to sync the peering with the current address space on the remote vNet after it's updated.
	SyncRemoteAddressSpace pulumi.StringPtrInput
	// Resource type.
	Type pulumi.StringPtrInput
	// If remote gateways can be used on this virtual network. If the flag is set to true, and allowGatewayTransit on remote peering is also true, virtual network will use gateways of remote virtual network for transit. Only one peering can have this flag set to true. This flag cannot be set if virtual network already has a gateway.
	UseRemoteGateways pulumi.BoolPtrInput
	// The name of the virtual network.
	VirtualNetworkName pulumi.StringInput
	// The name of the peering.
	VirtualNetworkPeeringName pulumi.StringPtrInput
}

func (VirtualNetworkPeeringArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*virtualNetworkPeeringArgs)(nil)).Elem()
}

type VirtualNetworkPeeringInput interface {
	pulumi.Input

	ToVirtualNetworkPeeringOutput() VirtualNetworkPeeringOutput
	ToVirtualNetworkPeeringOutputWithContext(ctx context.Context) VirtualNetworkPeeringOutput
}

func (*VirtualNetworkPeering) ElementType() reflect.Type {
	return reflect.TypeOf((**VirtualNetworkPeering)(nil)).Elem()
}

func (i *VirtualNetworkPeering) ToVirtualNetworkPeeringOutput() VirtualNetworkPeeringOutput {
	return i.ToVirtualNetworkPeeringOutputWithContext(context.Background())
}

func (i *VirtualNetworkPeering) ToVirtualNetworkPeeringOutputWithContext(ctx context.Context) VirtualNetworkPeeringOutput {
	return pulumi.ToOutputWithContext(ctx, i).(VirtualNetworkPeeringOutput)
}

type VirtualNetworkPeeringOutput struct{ *pulumi.OutputState }

func (VirtualNetworkPeeringOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**VirtualNetworkPeering)(nil)).Elem()
}

func (o VirtualNetworkPeeringOutput) ToVirtualNetworkPeeringOutput() VirtualNetworkPeeringOutput {
	return o
}

func (o VirtualNetworkPeeringOutput) ToVirtualNetworkPeeringOutputWithContext(ctx context.Context) VirtualNetworkPeeringOutput {
	return o
}

// Whether the forwarded traffic from the VMs in the local virtual network will be allowed/disallowed in remote virtual network.
func (o VirtualNetworkPeeringOutput) AllowForwardedTraffic() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.BoolPtrOutput { return v.AllowForwardedTraffic }).(pulumi.BoolPtrOutput)
}

// If gateway links can be used in remote virtual networking to link to this virtual network.
func (o VirtualNetworkPeeringOutput) AllowGatewayTransit() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.BoolPtrOutput { return v.AllowGatewayTransit }).(pulumi.BoolPtrOutput)
}

// Whether the VMs in the local virtual network space would be able to access the VMs in remote virtual network space.
func (o VirtualNetworkPeeringOutput) AllowVirtualNetworkAccess() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.BoolPtrOutput { return v.AllowVirtualNetworkAccess }).(pulumi.BoolPtrOutput)
}

// If we need to verify the provisioning state of the remote gateway.
func (o VirtualNetworkPeeringOutput) DoNotVerifyRemoteGateways() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.BoolPtrOutput { return v.DoNotVerifyRemoteGateways }).(pulumi.BoolPtrOutput)
}

// A unique read-only string that changes whenever the resource is updated.
func (o VirtualNetworkPeeringOutput) Etag() pulumi.StringOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringOutput { return v.Etag }).(pulumi.StringOutput)
}

// The name of the resource that is unique within a resource group. This name can be used to access the resource.
func (o VirtualNetworkPeeringOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringPtrOutput { return v.Name }).(pulumi.StringPtrOutput)
}

// The status of the virtual network peering.
func (o VirtualNetworkPeeringOutput) PeeringState() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringPtrOutput { return v.PeeringState }).(pulumi.StringPtrOutput)
}

// The peering sync status of the virtual network peering.
func (o VirtualNetworkPeeringOutput) PeeringSyncLevel() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringPtrOutput { return v.PeeringSyncLevel }).(pulumi.StringPtrOutput)
}

// The provisioning state of the virtual network peering resource.
func (o VirtualNetworkPeeringOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringOutput { return v.ProvisioningState }).(pulumi.StringOutput)
}

// The reference to the address space peered with the remote virtual network.
func (o VirtualNetworkPeeringOutput) RemoteAddressSpace() AddressSpaceResponsePtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) AddressSpaceResponsePtrOutput { return v.RemoteAddressSpace }).(AddressSpaceResponsePtrOutput)
}

// The reference to the remote virtual network's Bgp Communities.
func (o VirtualNetworkPeeringOutput) RemoteBgpCommunities() VirtualNetworkBgpCommunitiesResponsePtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) VirtualNetworkBgpCommunitiesResponsePtrOutput {
		return v.RemoteBgpCommunities
	}).(VirtualNetworkBgpCommunitiesResponsePtrOutput)
}

// The reference to the remote virtual network. The remote virtual network can be in the same or different region (preview). See here to register for the preview and learn more (https://docs.microsoft.com/en-us/azure/virtual-network/virtual-network-create-peering).
func (o VirtualNetworkPeeringOutput) RemoteVirtualNetwork() SubResourceResponsePtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) SubResourceResponsePtrOutput { return v.RemoteVirtualNetwork }).(SubResourceResponsePtrOutput)
}

// The reference to the current address space of the remote virtual network.
func (o VirtualNetworkPeeringOutput) RemoteVirtualNetworkAddressSpace() AddressSpaceResponsePtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) AddressSpaceResponsePtrOutput {
		return v.RemoteVirtualNetworkAddressSpace
	}).(AddressSpaceResponsePtrOutput)
}

// The reference to the remote virtual network's encryption
func (o VirtualNetworkPeeringOutput) RemoteVirtualNetworkEncryption() VirtualNetworkEncryptionResponseOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) VirtualNetworkEncryptionResponseOutput {
		return v.RemoteVirtualNetworkEncryption
	}).(VirtualNetworkEncryptionResponseOutput)
}

// The resourceGuid property of the Virtual Network peering resource.
func (o VirtualNetworkPeeringOutput) ResourceGuid() pulumi.StringOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringOutput { return v.ResourceGuid }).(pulumi.StringOutput)
}

// Resource type.
func (o VirtualNetworkPeeringOutput) Type() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.StringPtrOutput { return v.Type }).(pulumi.StringPtrOutput)
}

// If remote gateways can be used on this virtual network. If the flag is set to true, and allowGatewayTransit on remote peering is also true, virtual network will use gateways of remote virtual network for transit. Only one peering can have this flag set to true. This flag cannot be set if virtual network already has a gateway.
func (o VirtualNetworkPeeringOutput) UseRemoteGateways() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v *VirtualNetworkPeering) pulumi.BoolPtrOutput { return v.UseRemoteGateways }).(pulumi.BoolPtrOutput)
}

func init() {
	pulumi.RegisterOutputType(VirtualNetworkPeeringOutput{})
}
