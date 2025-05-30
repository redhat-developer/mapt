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

// Instance of Verifier Workspace.
//
// Uses Azure REST API version 2024-01-01-preview.
//
// Other available API versions: 2024-05-01.
type VerifierWorkspace struct {
	pulumi.CustomResourceState

	// The geo-location where the resource lives
	Location pulumi.StringOutput `pulumi:"location"`
	// The name of the resource
	Name pulumi.StringOutput `pulumi:"name"`
	// Properties of Verifier Workspace resource.
	Properties VerifierWorkspacePropertiesResponseOutput `pulumi:"properties"`
	// Azure Resource Manager metadata containing createdBy and modifiedBy information.
	SystemData SystemDataResponseOutput `pulumi:"systemData"`
	// Resource tags.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// The type of the resource. E.g. "Microsoft.Compute/virtualMachines" or "Microsoft.Storage/storageAccounts"
	Type pulumi.StringOutput `pulumi:"type"`
}

// NewVerifierWorkspace registers a new resource with the given unique name, arguments, and options.
func NewVerifierWorkspace(ctx *pulumi.Context,
	name string, args *VerifierWorkspaceArgs, opts ...pulumi.ResourceOption) (*VerifierWorkspace, error) {
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
			Type: pulumi.String("azure-native:network/v20240101preview:VerifierWorkspace"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240501:VerifierWorkspace"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource VerifierWorkspace
	err := ctx.RegisterResource("azure-native:network:VerifierWorkspace", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetVerifierWorkspace gets an existing VerifierWorkspace resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetVerifierWorkspace(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *VerifierWorkspaceState, opts ...pulumi.ResourceOption) (*VerifierWorkspace, error) {
	var resource VerifierWorkspace
	err := ctx.ReadResource("azure-native:network:VerifierWorkspace", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering VerifierWorkspace resources.
type verifierWorkspaceState struct {
}

type VerifierWorkspaceState struct {
}

func (VerifierWorkspaceState) ElementType() reflect.Type {
	return reflect.TypeOf((*verifierWorkspaceState)(nil)).Elem()
}

type verifierWorkspaceArgs struct {
	// The geo-location where the resource lives
	Location *string `pulumi:"location"`
	// The name of the network manager.
	NetworkManagerName string `pulumi:"networkManagerName"`
	// Properties of Verifier Workspace resource.
	Properties *VerifierWorkspaceProperties `pulumi:"properties"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// Resource tags.
	Tags map[string]string `pulumi:"tags"`
	// Workspace name.
	WorkspaceName *string `pulumi:"workspaceName"`
}

// The set of arguments for constructing a VerifierWorkspace resource.
type VerifierWorkspaceArgs struct {
	// The geo-location where the resource lives
	Location pulumi.StringPtrInput
	// The name of the network manager.
	NetworkManagerName pulumi.StringInput
	// Properties of Verifier Workspace resource.
	Properties VerifierWorkspacePropertiesPtrInput
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput
	// Resource tags.
	Tags pulumi.StringMapInput
	// Workspace name.
	WorkspaceName pulumi.StringPtrInput
}

func (VerifierWorkspaceArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*verifierWorkspaceArgs)(nil)).Elem()
}

type VerifierWorkspaceInput interface {
	pulumi.Input

	ToVerifierWorkspaceOutput() VerifierWorkspaceOutput
	ToVerifierWorkspaceOutputWithContext(ctx context.Context) VerifierWorkspaceOutput
}

func (*VerifierWorkspace) ElementType() reflect.Type {
	return reflect.TypeOf((**VerifierWorkspace)(nil)).Elem()
}

func (i *VerifierWorkspace) ToVerifierWorkspaceOutput() VerifierWorkspaceOutput {
	return i.ToVerifierWorkspaceOutputWithContext(context.Background())
}

func (i *VerifierWorkspace) ToVerifierWorkspaceOutputWithContext(ctx context.Context) VerifierWorkspaceOutput {
	return pulumi.ToOutputWithContext(ctx, i).(VerifierWorkspaceOutput)
}

type VerifierWorkspaceOutput struct{ *pulumi.OutputState }

func (VerifierWorkspaceOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**VerifierWorkspace)(nil)).Elem()
}

func (o VerifierWorkspaceOutput) ToVerifierWorkspaceOutput() VerifierWorkspaceOutput {
	return o
}

func (o VerifierWorkspaceOutput) ToVerifierWorkspaceOutputWithContext(ctx context.Context) VerifierWorkspaceOutput {
	return o
}

// The geo-location where the resource lives
func (o VerifierWorkspaceOutput) Location() pulumi.StringOutput {
	return o.ApplyT(func(v *VerifierWorkspace) pulumi.StringOutput { return v.Location }).(pulumi.StringOutput)
}

// The name of the resource
func (o VerifierWorkspaceOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *VerifierWorkspace) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// Properties of Verifier Workspace resource.
func (o VerifierWorkspaceOutput) Properties() VerifierWorkspacePropertiesResponseOutput {
	return o.ApplyT(func(v *VerifierWorkspace) VerifierWorkspacePropertiesResponseOutput { return v.Properties }).(VerifierWorkspacePropertiesResponseOutput)
}

// Azure Resource Manager metadata containing createdBy and modifiedBy information.
func (o VerifierWorkspaceOutput) SystemData() SystemDataResponseOutput {
	return o.ApplyT(func(v *VerifierWorkspace) SystemDataResponseOutput { return v.SystemData }).(SystemDataResponseOutput)
}

// Resource tags.
func (o VerifierWorkspaceOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *VerifierWorkspace) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// The type of the resource. E.g. "Microsoft.Compute/virtualMachines" or "Microsoft.Storage/storageAccounts"
func (o VerifierWorkspaceOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *VerifierWorkspace) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(VerifierWorkspaceOutput{})
}
