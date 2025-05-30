// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package compute

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Specifies information about the gallery inVMAccessControlProfile that you want to create or update.
//
// Uses Azure REST API version 2024-03-03.
type GalleryInVMAccessControlProfile struct {
	pulumi.CustomResourceState

	// Resource location
	Location pulumi.StringOutput `pulumi:"location"`
	// Resource name
	Name pulumi.StringOutput `pulumi:"name"`
	// Describes the properties of a gallery inVMAccessControlProfile.
	Properties GalleryInVMAccessControlProfilePropertiesResponseOutput `pulumi:"properties"`
	// Resource tags
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// Resource type
	Type pulumi.StringOutput `pulumi:"type"`
}

// NewGalleryInVMAccessControlProfile registers a new resource with the given unique name, arguments, and options.
func NewGalleryInVMAccessControlProfile(ctx *pulumi.Context,
	name string, args *GalleryInVMAccessControlProfileArgs, opts ...pulumi.ResourceOption) (*GalleryInVMAccessControlProfile, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.GalleryName == nil {
		return nil, errors.New("invalid value for required argument 'GalleryName'")
	}
	if args.ResourceGroupName == nil {
		return nil, errors.New("invalid value for required argument 'ResourceGroupName'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("azure-native:compute/v20240303:GalleryInVMAccessControlProfile"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource GalleryInVMAccessControlProfile
	err := ctx.RegisterResource("azure-native:compute:GalleryInVMAccessControlProfile", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetGalleryInVMAccessControlProfile gets an existing GalleryInVMAccessControlProfile resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetGalleryInVMAccessControlProfile(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *GalleryInVMAccessControlProfileState, opts ...pulumi.ResourceOption) (*GalleryInVMAccessControlProfile, error) {
	var resource GalleryInVMAccessControlProfile
	err := ctx.ReadResource("azure-native:compute:GalleryInVMAccessControlProfile", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering GalleryInVMAccessControlProfile resources.
type galleryInVMAccessControlProfileState struct {
}

type GalleryInVMAccessControlProfileState struct {
}

func (GalleryInVMAccessControlProfileState) ElementType() reflect.Type {
	return reflect.TypeOf((*galleryInVMAccessControlProfileState)(nil)).Elem()
}

type galleryInVMAccessControlProfileArgs struct {
	// The name of the Shared Image Gallery in which the InVMAccessControlProfile is to be created.
	GalleryName string `pulumi:"galleryName"`
	// The name of the gallery inVMAccessControlProfile to be created or updated. The allowed characters are alphabets and numbers with dots, dashes, and periods allowed in the middle. The maximum length is 80 characters.
	InVMAccessControlProfileName *string `pulumi:"inVMAccessControlProfileName"`
	// Resource location
	Location *string `pulumi:"location"`
	// Describes the properties of a gallery inVMAccessControlProfile.
	Properties *GalleryInVMAccessControlProfileProperties `pulumi:"properties"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// Resource tags
	Tags map[string]string `pulumi:"tags"`
}

// The set of arguments for constructing a GalleryInVMAccessControlProfile resource.
type GalleryInVMAccessControlProfileArgs struct {
	// The name of the Shared Image Gallery in which the InVMAccessControlProfile is to be created.
	GalleryName pulumi.StringInput
	// The name of the gallery inVMAccessControlProfile to be created or updated. The allowed characters are alphabets and numbers with dots, dashes, and periods allowed in the middle. The maximum length is 80 characters.
	InVMAccessControlProfileName pulumi.StringPtrInput
	// Resource location
	Location pulumi.StringPtrInput
	// Describes the properties of a gallery inVMAccessControlProfile.
	Properties GalleryInVMAccessControlProfilePropertiesPtrInput
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput
	// Resource tags
	Tags pulumi.StringMapInput
}

func (GalleryInVMAccessControlProfileArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*galleryInVMAccessControlProfileArgs)(nil)).Elem()
}

type GalleryInVMAccessControlProfileInput interface {
	pulumi.Input

	ToGalleryInVMAccessControlProfileOutput() GalleryInVMAccessControlProfileOutput
	ToGalleryInVMAccessControlProfileOutputWithContext(ctx context.Context) GalleryInVMAccessControlProfileOutput
}

func (*GalleryInVMAccessControlProfile) ElementType() reflect.Type {
	return reflect.TypeOf((**GalleryInVMAccessControlProfile)(nil)).Elem()
}

func (i *GalleryInVMAccessControlProfile) ToGalleryInVMAccessControlProfileOutput() GalleryInVMAccessControlProfileOutput {
	return i.ToGalleryInVMAccessControlProfileOutputWithContext(context.Background())
}

func (i *GalleryInVMAccessControlProfile) ToGalleryInVMAccessControlProfileOutputWithContext(ctx context.Context) GalleryInVMAccessControlProfileOutput {
	return pulumi.ToOutputWithContext(ctx, i).(GalleryInVMAccessControlProfileOutput)
}

type GalleryInVMAccessControlProfileOutput struct{ *pulumi.OutputState }

func (GalleryInVMAccessControlProfileOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**GalleryInVMAccessControlProfile)(nil)).Elem()
}

func (o GalleryInVMAccessControlProfileOutput) ToGalleryInVMAccessControlProfileOutput() GalleryInVMAccessControlProfileOutput {
	return o
}

func (o GalleryInVMAccessControlProfileOutput) ToGalleryInVMAccessControlProfileOutputWithContext(ctx context.Context) GalleryInVMAccessControlProfileOutput {
	return o
}

// Resource location
func (o GalleryInVMAccessControlProfileOutput) Location() pulumi.StringOutput {
	return o.ApplyT(func(v *GalleryInVMAccessControlProfile) pulumi.StringOutput { return v.Location }).(pulumi.StringOutput)
}

// Resource name
func (o GalleryInVMAccessControlProfileOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *GalleryInVMAccessControlProfile) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// Describes the properties of a gallery inVMAccessControlProfile.
func (o GalleryInVMAccessControlProfileOutput) Properties() GalleryInVMAccessControlProfilePropertiesResponseOutput {
	return o.ApplyT(func(v *GalleryInVMAccessControlProfile) GalleryInVMAccessControlProfilePropertiesResponseOutput {
		return v.Properties
	}).(GalleryInVMAccessControlProfilePropertiesResponseOutput)
}

// Resource tags
func (o GalleryInVMAccessControlProfileOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *GalleryInVMAccessControlProfile) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// Resource type
func (o GalleryInVMAccessControlProfileOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *GalleryInVMAccessControlProfile) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(GalleryInVMAccessControlProfileOutput{})
}
