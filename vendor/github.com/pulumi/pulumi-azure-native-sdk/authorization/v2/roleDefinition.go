// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package authorization

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Role definition.
//
// Uses Azure REST API version 2022-05-01-preview. In version 1.x of the Azure Native provider, it used API version 2018-01-01-preview.
type RoleDefinition struct {
	pulumi.CustomResourceState

	// Role definition assignable scopes.
	AssignableScopes pulumi.StringArrayOutput `pulumi:"assignableScopes"`
	// Id of the user who created the assignment
	CreatedBy pulumi.StringOutput `pulumi:"createdBy"`
	// Time it was created
	CreatedOn pulumi.StringOutput `pulumi:"createdOn"`
	// The role definition description.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// The role definition name.
	Name pulumi.StringOutput `pulumi:"name"`
	// Role definition permissions.
	Permissions PermissionResponseArrayOutput `pulumi:"permissions"`
	// The role name.
	RoleName pulumi.StringPtrOutput `pulumi:"roleName"`
	// The role type.
	RoleType pulumi.StringPtrOutput `pulumi:"roleType"`
	// The role definition type.
	Type pulumi.StringOutput `pulumi:"type"`
	// Id of the user who updated the assignment
	UpdatedBy pulumi.StringOutput `pulumi:"updatedBy"`
	// Time it was updated
	UpdatedOn pulumi.StringOutput `pulumi:"updatedOn"`
}

// NewRoleDefinition registers a new resource with the given unique name, arguments, and options.
func NewRoleDefinition(ctx *pulumi.Context,
	name string, args *RoleDefinitionArgs, opts ...pulumi.ResourceOption) (*RoleDefinition, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Scope == nil {
		return nil, errors.New("invalid value for required argument 'Scope'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("azure-native:authorization/v20150701:RoleDefinition"),
		},
		{
			Type: pulumi.String("azure-native:authorization/v20180101preview:RoleDefinition"),
		},
		{
			Type: pulumi.String("azure-native:authorization/v20220401:RoleDefinition"),
		},
		{
			Type: pulumi.String("azure-native:authorization/v20220501preview:RoleDefinition"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource RoleDefinition
	err := ctx.RegisterResource("azure-native:authorization:RoleDefinition", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetRoleDefinition gets an existing RoleDefinition resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetRoleDefinition(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *RoleDefinitionState, opts ...pulumi.ResourceOption) (*RoleDefinition, error) {
	var resource RoleDefinition
	err := ctx.ReadResource("azure-native:authorization:RoleDefinition", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering RoleDefinition resources.
type roleDefinitionState struct {
}

type RoleDefinitionState struct {
}

func (RoleDefinitionState) ElementType() reflect.Type {
	return reflect.TypeOf((*roleDefinitionState)(nil)).Elem()
}

type roleDefinitionArgs struct {
	// Role definition assignable scopes.
	AssignableScopes []string `pulumi:"assignableScopes"`
	// The role definition description.
	Description *string `pulumi:"description"`
	// Role definition permissions.
	Permissions []Permission `pulumi:"permissions"`
	// The ID of the role definition.
	RoleDefinitionId *string `pulumi:"roleDefinitionId"`
	// The role name.
	RoleName *string `pulumi:"roleName"`
	// The role type.
	RoleType *string `pulumi:"roleType"`
	// The scope of the operation or resource. Valid scopes are: subscription (format: '/subscriptions/{subscriptionId}'), resource group (format: '/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}', or resource (format: '/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/[{parentResourcePath}/]{resourceType}/{resourceName}'
	Scope string `pulumi:"scope"`
}

// The set of arguments for constructing a RoleDefinition resource.
type RoleDefinitionArgs struct {
	// Role definition assignable scopes.
	AssignableScopes pulumi.StringArrayInput
	// The role definition description.
	Description pulumi.StringPtrInput
	// Role definition permissions.
	Permissions PermissionArrayInput
	// The ID of the role definition.
	RoleDefinitionId pulumi.StringPtrInput
	// The role name.
	RoleName pulumi.StringPtrInput
	// The role type.
	RoleType pulumi.StringPtrInput
	// The scope of the operation or resource. Valid scopes are: subscription (format: '/subscriptions/{subscriptionId}'), resource group (format: '/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}', or resource (format: '/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/[{parentResourcePath}/]{resourceType}/{resourceName}'
	Scope pulumi.StringInput
}

func (RoleDefinitionArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*roleDefinitionArgs)(nil)).Elem()
}

type RoleDefinitionInput interface {
	pulumi.Input

	ToRoleDefinitionOutput() RoleDefinitionOutput
	ToRoleDefinitionOutputWithContext(ctx context.Context) RoleDefinitionOutput
}

func (*RoleDefinition) ElementType() reflect.Type {
	return reflect.TypeOf((**RoleDefinition)(nil)).Elem()
}

func (i *RoleDefinition) ToRoleDefinitionOutput() RoleDefinitionOutput {
	return i.ToRoleDefinitionOutputWithContext(context.Background())
}

func (i *RoleDefinition) ToRoleDefinitionOutputWithContext(ctx context.Context) RoleDefinitionOutput {
	return pulumi.ToOutputWithContext(ctx, i).(RoleDefinitionOutput)
}

type RoleDefinitionOutput struct{ *pulumi.OutputState }

func (RoleDefinitionOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**RoleDefinition)(nil)).Elem()
}

func (o RoleDefinitionOutput) ToRoleDefinitionOutput() RoleDefinitionOutput {
	return o
}

func (o RoleDefinitionOutput) ToRoleDefinitionOutputWithContext(ctx context.Context) RoleDefinitionOutput {
	return o
}

// Role definition assignable scopes.
func (o RoleDefinitionOutput) AssignableScopes() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringArrayOutput { return v.AssignableScopes }).(pulumi.StringArrayOutput)
}

// Id of the user who created the assignment
func (o RoleDefinitionOutput) CreatedBy() pulumi.StringOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringOutput { return v.CreatedBy }).(pulumi.StringOutput)
}

// Time it was created
func (o RoleDefinitionOutput) CreatedOn() pulumi.StringOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringOutput { return v.CreatedOn }).(pulumi.StringOutput)
}

// The role definition description.
func (o RoleDefinitionOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// The role definition name.
func (o RoleDefinitionOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// Role definition permissions.
func (o RoleDefinitionOutput) Permissions() PermissionResponseArrayOutput {
	return o.ApplyT(func(v *RoleDefinition) PermissionResponseArrayOutput { return v.Permissions }).(PermissionResponseArrayOutput)
}

// The role name.
func (o RoleDefinitionOutput) RoleName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringPtrOutput { return v.RoleName }).(pulumi.StringPtrOutput)
}

// The role type.
func (o RoleDefinitionOutput) RoleType() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringPtrOutput { return v.RoleType }).(pulumi.StringPtrOutput)
}

// The role definition type.
func (o RoleDefinitionOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

// Id of the user who updated the assignment
func (o RoleDefinitionOutput) UpdatedBy() pulumi.StringOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringOutput { return v.UpdatedBy }).(pulumi.StringOutput)
}

// Time it was updated
func (o RoleDefinitionOutput) UpdatedOn() pulumi.StringOutput {
	return o.ApplyT(func(v *RoleDefinition) pulumi.StringOutput { return v.UpdatedOn }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(RoleDefinitionOutput{})
}
