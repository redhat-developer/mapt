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

// The Network Security Perimeter resource
//
// Uses Azure REST API version 2021-03-01-preview. In version 1.x of the Azure Native provider, it used API version 2021-02-01-preview.
//
// Other available API versions: 2021-02-01-preview, 2023-07-01-preview, 2023-08-01-preview, 2024-06-01-preview.
type NetworkSecurityPerimeter struct {
	pulumi.CustomResourceState

	// A description of the network security perimeter.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// A friendly name for the network security perimeter.
	DisplayName pulumi.StringPtrOutput `pulumi:"displayName"`
	// A unique read-only string that changes whenever the resource is updated.
	Etag pulumi.StringOutput `pulumi:"etag"`
	// Resource location.
	Location pulumi.StringPtrOutput `pulumi:"location"`
	// Resource name.
	Name pulumi.StringOutput `pulumi:"name"`
	// The provisioning state of the scope assignment resource.
	ProvisioningState pulumi.StringOutput `pulumi:"provisioningState"`
	// Resource tags.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// Resource type.
	Type pulumi.StringOutput `pulumi:"type"`
}

// NewNetworkSecurityPerimeter registers a new resource with the given unique name, arguments, and options.
func NewNetworkSecurityPerimeter(ctx *pulumi.Context,
	name string, args *NetworkSecurityPerimeterArgs, opts ...pulumi.ResourceOption) (*NetworkSecurityPerimeter, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.ResourceGroupName == nil {
		return nil, errors.New("invalid value for required argument 'ResourceGroupName'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("azure-native:network/v20210201preview:NetworkSecurityPerimeter"),
		},
		{
			Type: pulumi.String("azure-native:network/v20210301preview:NetworkSecurityPerimeter"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230701preview:NetworkSecurityPerimeter"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230801preview:NetworkSecurityPerimeter"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240601preview:NetworkSecurityPerimeter"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource NetworkSecurityPerimeter
	err := ctx.RegisterResource("azure-native:network:NetworkSecurityPerimeter", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetNetworkSecurityPerimeter gets an existing NetworkSecurityPerimeter resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetNetworkSecurityPerimeter(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *NetworkSecurityPerimeterState, opts ...pulumi.ResourceOption) (*NetworkSecurityPerimeter, error) {
	var resource NetworkSecurityPerimeter
	err := ctx.ReadResource("azure-native:network:NetworkSecurityPerimeter", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering NetworkSecurityPerimeter resources.
type networkSecurityPerimeterState struct {
}

type NetworkSecurityPerimeterState struct {
}

func (NetworkSecurityPerimeterState) ElementType() reflect.Type {
	return reflect.TypeOf((*networkSecurityPerimeterState)(nil)).Elem()
}

type networkSecurityPerimeterArgs struct {
	// A description of the network security perimeter.
	Description *string `pulumi:"description"`
	// A friendly name for the network security perimeter.
	DisplayName *string `pulumi:"displayName"`
	// Resource ID.
	Id *string `pulumi:"id"`
	// Resource location.
	Location *string `pulumi:"location"`
	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `pulumi:"name"`
	// The name of the network security perimeter.
	NetworkSecurityPerimeterName *string `pulumi:"networkSecurityPerimeterName"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// Resource tags.
	Tags map[string]string `pulumi:"tags"`
}

// The set of arguments for constructing a NetworkSecurityPerimeter resource.
type NetworkSecurityPerimeterArgs struct {
	// A description of the network security perimeter.
	Description pulumi.StringPtrInput
	// A friendly name for the network security perimeter.
	DisplayName pulumi.StringPtrInput
	// Resource ID.
	Id pulumi.StringPtrInput
	// Resource location.
	Location pulumi.StringPtrInput
	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name pulumi.StringPtrInput
	// The name of the network security perimeter.
	NetworkSecurityPerimeterName pulumi.StringPtrInput
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput
	// Resource tags.
	Tags pulumi.StringMapInput
}

func (NetworkSecurityPerimeterArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*networkSecurityPerimeterArgs)(nil)).Elem()
}

type NetworkSecurityPerimeterInput interface {
	pulumi.Input

	ToNetworkSecurityPerimeterOutput() NetworkSecurityPerimeterOutput
	ToNetworkSecurityPerimeterOutputWithContext(ctx context.Context) NetworkSecurityPerimeterOutput
}

func (*NetworkSecurityPerimeter) ElementType() reflect.Type {
	return reflect.TypeOf((**NetworkSecurityPerimeter)(nil)).Elem()
}

func (i *NetworkSecurityPerimeter) ToNetworkSecurityPerimeterOutput() NetworkSecurityPerimeterOutput {
	return i.ToNetworkSecurityPerimeterOutputWithContext(context.Background())
}

func (i *NetworkSecurityPerimeter) ToNetworkSecurityPerimeterOutputWithContext(ctx context.Context) NetworkSecurityPerimeterOutput {
	return pulumi.ToOutputWithContext(ctx, i).(NetworkSecurityPerimeterOutput)
}

type NetworkSecurityPerimeterOutput struct{ *pulumi.OutputState }

func (NetworkSecurityPerimeterOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**NetworkSecurityPerimeter)(nil)).Elem()
}

func (o NetworkSecurityPerimeterOutput) ToNetworkSecurityPerimeterOutput() NetworkSecurityPerimeterOutput {
	return o
}

func (o NetworkSecurityPerimeterOutput) ToNetworkSecurityPerimeterOutputWithContext(ctx context.Context) NetworkSecurityPerimeterOutput {
	return o
}

// A description of the network security perimeter.
func (o NetworkSecurityPerimeterOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// A friendly name for the network security perimeter.
func (o NetworkSecurityPerimeterOutput) DisplayName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringPtrOutput { return v.DisplayName }).(pulumi.StringPtrOutput)
}

// A unique read-only string that changes whenever the resource is updated.
func (o NetworkSecurityPerimeterOutput) Etag() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringOutput { return v.Etag }).(pulumi.StringOutput)
}

// Resource location.
func (o NetworkSecurityPerimeterOutput) Location() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringPtrOutput { return v.Location }).(pulumi.StringPtrOutput)
}

// Resource name.
func (o NetworkSecurityPerimeterOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// The provisioning state of the scope assignment resource.
func (o NetworkSecurityPerimeterOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringOutput { return v.ProvisioningState }).(pulumi.StringOutput)
}

// Resource tags.
func (o NetworkSecurityPerimeterOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// Resource type.
func (o NetworkSecurityPerimeterOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkSecurityPerimeter) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(NetworkSecurityPerimeterOutput{})
}
