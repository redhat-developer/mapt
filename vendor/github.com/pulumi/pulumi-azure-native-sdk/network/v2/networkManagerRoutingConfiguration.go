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

// Defines the routing configuration
//
// Uses Azure REST API version 2024-03-01.
//
// Other available API versions: 2024-05-01.
type NetworkManagerRoutingConfiguration struct {
	pulumi.CustomResourceState

	// A description of the routing configuration.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// A unique read-only string that changes whenever the resource is updated.
	Etag pulumi.StringOutput `pulumi:"etag"`
	// Resource name.
	Name pulumi.StringOutput `pulumi:"name"`
	// The provisioning state of the resource.
	ProvisioningState pulumi.StringOutput `pulumi:"provisioningState"`
	// Unique identifier for this resource.
	ResourceGuid pulumi.StringOutput `pulumi:"resourceGuid"`
	// The system metadata related to this resource.
	SystemData SystemDataResponseOutput `pulumi:"systemData"`
	// Resource type.
	Type pulumi.StringOutput `pulumi:"type"`
}

// NewNetworkManagerRoutingConfiguration registers a new resource with the given unique name, arguments, and options.
func NewNetworkManagerRoutingConfiguration(ctx *pulumi.Context,
	name string, args *NetworkManagerRoutingConfigurationArgs, opts ...pulumi.ResourceOption) (*NetworkManagerRoutingConfiguration, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.NetworkManagerName == nil {
		return nil, errors.New("invalid value for required argument 'NetworkManagerName'")
	}
	if args.ResourceGroupName == nil {
		return nil, errors.New("invalid value for required argument 'ResourceGroupName'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("azure-native:network/v20240301:NetworkManagerRoutingConfiguration"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240501:NetworkManagerRoutingConfiguration"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource NetworkManagerRoutingConfiguration
	err := ctx.RegisterResource("azure-native:network:NetworkManagerRoutingConfiguration", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetNetworkManagerRoutingConfiguration gets an existing NetworkManagerRoutingConfiguration resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetNetworkManagerRoutingConfiguration(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *NetworkManagerRoutingConfigurationState, opts ...pulumi.ResourceOption) (*NetworkManagerRoutingConfiguration, error) {
	var resource NetworkManagerRoutingConfiguration
	err := ctx.ReadResource("azure-native:network:NetworkManagerRoutingConfiguration", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering NetworkManagerRoutingConfiguration resources.
type networkManagerRoutingConfigurationState struct {
}

type NetworkManagerRoutingConfigurationState struct {
}

func (NetworkManagerRoutingConfigurationState) ElementType() reflect.Type {
	return reflect.TypeOf((*networkManagerRoutingConfigurationState)(nil)).Elem()
}

type networkManagerRoutingConfigurationArgs struct {
	// The name of the network manager Routing Configuration.
	ConfigurationName *string `pulumi:"configurationName"`
	// A description of the routing configuration.
	Description *string `pulumi:"description"`
	// The name of the network manager.
	NetworkManagerName string `pulumi:"networkManagerName"`
	// The name of the resource group. The name is case insensitive.
	ResourceGroupName string `pulumi:"resourceGroupName"`
}

// The set of arguments for constructing a NetworkManagerRoutingConfiguration resource.
type NetworkManagerRoutingConfigurationArgs struct {
	// The name of the network manager Routing Configuration.
	ConfigurationName pulumi.StringPtrInput
	// A description of the routing configuration.
	Description pulumi.StringPtrInput
	// The name of the network manager.
	NetworkManagerName pulumi.StringInput
	// The name of the resource group. The name is case insensitive.
	ResourceGroupName pulumi.StringInput
}

func (NetworkManagerRoutingConfigurationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*networkManagerRoutingConfigurationArgs)(nil)).Elem()
}

type NetworkManagerRoutingConfigurationInput interface {
	pulumi.Input

	ToNetworkManagerRoutingConfigurationOutput() NetworkManagerRoutingConfigurationOutput
	ToNetworkManagerRoutingConfigurationOutputWithContext(ctx context.Context) NetworkManagerRoutingConfigurationOutput
}

func (*NetworkManagerRoutingConfiguration) ElementType() reflect.Type {
	return reflect.TypeOf((**NetworkManagerRoutingConfiguration)(nil)).Elem()
}

func (i *NetworkManagerRoutingConfiguration) ToNetworkManagerRoutingConfigurationOutput() NetworkManagerRoutingConfigurationOutput {
	return i.ToNetworkManagerRoutingConfigurationOutputWithContext(context.Background())
}

func (i *NetworkManagerRoutingConfiguration) ToNetworkManagerRoutingConfigurationOutputWithContext(ctx context.Context) NetworkManagerRoutingConfigurationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(NetworkManagerRoutingConfigurationOutput)
}

type NetworkManagerRoutingConfigurationOutput struct{ *pulumi.OutputState }

func (NetworkManagerRoutingConfigurationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**NetworkManagerRoutingConfiguration)(nil)).Elem()
}

func (o NetworkManagerRoutingConfigurationOutput) ToNetworkManagerRoutingConfigurationOutput() NetworkManagerRoutingConfigurationOutput {
	return o
}

func (o NetworkManagerRoutingConfigurationOutput) ToNetworkManagerRoutingConfigurationOutputWithContext(ctx context.Context) NetworkManagerRoutingConfigurationOutput {
	return o
}

// A description of the routing configuration.
func (o NetworkManagerRoutingConfigurationOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// A unique read-only string that changes whenever the resource is updated.
func (o NetworkManagerRoutingConfigurationOutput) Etag() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) pulumi.StringOutput { return v.Etag }).(pulumi.StringOutput)
}

// Resource name.
func (o NetworkManagerRoutingConfigurationOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// The provisioning state of the resource.
func (o NetworkManagerRoutingConfigurationOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) pulumi.StringOutput { return v.ProvisioningState }).(pulumi.StringOutput)
}

// Unique identifier for this resource.
func (o NetworkManagerRoutingConfigurationOutput) ResourceGuid() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) pulumi.StringOutput { return v.ResourceGuid }).(pulumi.StringOutput)
}

// The system metadata related to this resource.
func (o NetworkManagerRoutingConfigurationOutput) SystemData() SystemDataResponseOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) SystemDataResponseOutput { return v.SystemData }).(SystemDataResponseOutput)
}

// Resource type.
func (o NetworkManagerRoutingConfigurationOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkManagerRoutingConfiguration) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(NetworkManagerRoutingConfigurationOutput{})
}
