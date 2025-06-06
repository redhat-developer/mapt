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

// Network admin rule.
//
// Uses Azure REST API version 2023-02-01. In version 1.x of the Azure Native provider, it used API version 2021-02-01-preview.
//
// Other available API versions: 2021-02-01-preview, 2021-05-01-preview, 2023-04-01, 2023-05-01, 2023-06-01, 2023-09-01, 2023-11-01, 2024-01-01, 2024-01-01-preview, 2024-03-01, 2024-05-01.
type AdminRule struct {
	pulumi.CustomResourceState

	// Indicates the access allowed for this particular rule
	Access pulumi.StringOutput `pulumi:"access"`
	// A description for this rule. Restricted to 140 chars.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// The destination port ranges.
	DestinationPortRanges pulumi.StringArrayOutput `pulumi:"destinationPortRanges"`
	// The destination address prefixes. CIDR or destination IP ranges.
	Destinations AddressPrefixItemResponseArrayOutput `pulumi:"destinations"`
	// Indicates if the traffic matched against the rule in inbound or outbound.
	Direction pulumi.StringOutput `pulumi:"direction"`
	// A unique read-only string that changes whenever the resource is updated.
	Etag pulumi.StringOutput `pulumi:"etag"`
	// Whether the rule is custom or default.
	// Expected value is 'Custom'.
	Kind pulumi.StringOutput `pulumi:"kind"`
	// Resource name.
	Name pulumi.StringOutput `pulumi:"name"`
	// The priority of the rule. The value can be between 1 and 4096. The priority number must be unique for each rule in the collection. The lower the priority number, the higher the priority of the rule.
	Priority pulumi.IntOutput `pulumi:"priority"`
	// Network protocol this rule applies to.
	Protocol pulumi.StringOutput `pulumi:"protocol"`
	// The provisioning state of the resource.
	ProvisioningState pulumi.StringOutput `pulumi:"provisioningState"`
	// Unique identifier for this resource.
	ResourceGuid pulumi.StringOutput `pulumi:"resourceGuid"`
	// The source port ranges.
	SourcePortRanges pulumi.StringArrayOutput `pulumi:"sourcePortRanges"`
	// The CIDR or source IP ranges.
	Sources AddressPrefixItemResponseArrayOutput `pulumi:"sources"`
	// The system metadata related to this resource.
	SystemData SystemDataResponseOutput `pulumi:"systemData"`
	// Resource type.
	Type pulumi.StringOutput `pulumi:"type"`
}

// NewAdminRule registers a new resource with the given unique name, arguments, and options.
func NewAdminRule(ctx *pulumi.Context,
	name string, args *AdminRuleArgs, opts ...pulumi.ResourceOption) (*AdminRule, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Access == nil {
		return nil, errors.New("invalid value for required argument 'Access'")
	}
	if args.ConfigurationName == nil {
		return nil, errors.New("invalid value for required argument 'ConfigurationName'")
	}
	if args.Direction == nil {
		return nil, errors.New("invalid value for required argument 'Direction'")
	}
	if args.Kind == nil {
		return nil, errors.New("invalid value for required argument 'Kind'")
	}
	if args.NetworkManagerName == nil {
		return nil, errors.New("invalid value for required argument 'NetworkManagerName'")
	}
	if args.Priority == nil {
		return nil, errors.New("invalid value for required argument 'Priority'")
	}
	if args.Protocol == nil {
		return nil, errors.New("invalid value for required argument 'Protocol'")
	}
	if args.ResourceGroupName == nil {
		return nil, errors.New("invalid value for required argument 'ResourceGroupName'")
	}
	if args.RuleCollectionName == nil {
		return nil, errors.New("invalid value for required argument 'RuleCollectionName'")
	}
	args.Kind = pulumi.String("Custom")
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("azure-native:network/v20210201preview:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20210501preview:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220101:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220201preview:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220401preview:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220501:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220701:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20220901:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20221101:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230201:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230401:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230501:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230601:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20230901:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20231101:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240101:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240101preview:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240301:AdminRule"),
		},
		{
			Type: pulumi.String("azure-native:network/v20240501:AdminRule"),
		},
	})
	opts = append(opts, aliases)
	opts = utilities.PkgResourceDefaultOpts(opts)
	var resource AdminRule
	err := ctx.RegisterResource("azure-native:network:AdminRule", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetAdminRule gets an existing AdminRule resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetAdminRule(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *AdminRuleState, opts ...pulumi.ResourceOption) (*AdminRule, error) {
	var resource AdminRule
	err := ctx.ReadResource("azure-native:network:AdminRule", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering AdminRule resources.
type adminRuleState struct {
}

type AdminRuleState struct {
}

func (AdminRuleState) ElementType() reflect.Type {
	return reflect.TypeOf((*adminRuleState)(nil)).Elem()
}

type adminRuleArgs struct {
	// Indicates the access allowed for this particular rule
	Access string `pulumi:"access"`
	// The name of the network manager Security Configuration.
	ConfigurationName string `pulumi:"configurationName"`
	// A description for this rule. Restricted to 140 chars.
	Description *string `pulumi:"description"`
	// The destination port ranges.
	DestinationPortRanges []string `pulumi:"destinationPortRanges"`
	// The destination address prefixes. CIDR or destination IP ranges.
	Destinations []AddressPrefixItem `pulumi:"destinations"`
	// Indicates if the traffic matched against the rule in inbound or outbound.
	Direction string `pulumi:"direction"`
	// Whether the rule is custom or default.
	// Expected value is 'Custom'.
	Kind string `pulumi:"kind"`
	// The name of the network manager.
	NetworkManagerName string `pulumi:"networkManagerName"`
	// The priority of the rule. The value can be between 1 and 4096. The priority number must be unique for each rule in the collection. The lower the priority number, the higher the priority of the rule.
	Priority int `pulumi:"priority"`
	// Network protocol this rule applies to.
	Protocol string `pulumi:"protocol"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// The name of the network manager security Configuration rule collection.
	RuleCollectionName string `pulumi:"ruleCollectionName"`
	// The name of the rule.
	RuleName *string `pulumi:"ruleName"`
	// The source port ranges.
	SourcePortRanges []string `pulumi:"sourcePortRanges"`
	// The CIDR or source IP ranges.
	Sources []AddressPrefixItem `pulumi:"sources"`
}

// The set of arguments for constructing a AdminRule resource.
type AdminRuleArgs struct {
	// Indicates the access allowed for this particular rule
	Access pulumi.StringInput
	// The name of the network manager Security Configuration.
	ConfigurationName pulumi.StringInput
	// A description for this rule. Restricted to 140 chars.
	Description pulumi.StringPtrInput
	// The destination port ranges.
	DestinationPortRanges pulumi.StringArrayInput
	// The destination address prefixes. CIDR or destination IP ranges.
	Destinations AddressPrefixItemArrayInput
	// Indicates if the traffic matched against the rule in inbound or outbound.
	Direction pulumi.StringInput
	// Whether the rule is custom or default.
	// Expected value is 'Custom'.
	Kind pulumi.StringInput
	// The name of the network manager.
	NetworkManagerName pulumi.StringInput
	// The priority of the rule. The value can be between 1 and 4096. The priority number must be unique for each rule in the collection. The lower the priority number, the higher the priority of the rule.
	Priority pulumi.IntInput
	// Network protocol this rule applies to.
	Protocol pulumi.StringInput
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput
	// The name of the network manager security Configuration rule collection.
	RuleCollectionName pulumi.StringInput
	// The name of the rule.
	RuleName pulumi.StringPtrInput
	// The source port ranges.
	SourcePortRanges pulumi.StringArrayInput
	// The CIDR or source IP ranges.
	Sources AddressPrefixItemArrayInput
}

func (AdminRuleArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*adminRuleArgs)(nil)).Elem()
}

type AdminRuleInput interface {
	pulumi.Input

	ToAdminRuleOutput() AdminRuleOutput
	ToAdminRuleOutputWithContext(ctx context.Context) AdminRuleOutput
}

func (*AdminRule) ElementType() reflect.Type {
	return reflect.TypeOf((**AdminRule)(nil)).Elem()
}

func (i *AdminRule) ToAdminRuleOutput() AdminRuleOutput {
	return i.ToAdminRuleOutputWithContext(context.Background())
}

func (i *AdminRule) ToAdminRuleOutputWithContext(ctx context.Context) AdminRuleOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AdminRuleOutput)
}

type AdminRuleOutput struct{ *pulumi.OutputState }

func (AdminRuleOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**AdminRule)(nil)).Elem()
}

func (o AdminRuleOutput) ToAdminRuleOutput() AdminRuleOutput {
	return o
}

func (o AdminRuleOutput) ToAdminRuleOutputWithContext(ctx context.Context) AdminRuleOutput {
	return o
}

// Indicates the access allowed for this particular rule
func (o AdminRuleOutput) Access() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Access }).(pulumi.StringOutput)
}

// A description for this rule. Restricted to 140 chars.
func (o AdminRuleOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// The destination port ranges.
func (o AdminRuleOutput) DestinationPortRanges() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringArrayOutput { return v.DestinationPortRanges }).(pulumi.StringArrayOutput)
}

// The destination address prefixes. CIDR or destination IP ranges.
func (o AdminRuleOutput) Destinations() AddressPrefixItemResponseArrayOutput {
	return o.ApplyT(func(v *AdminRule) AddressPrefixItemResponseArrayOutput { return v.Destinations }).(AddressPrefixItemResponseArrayOutput)
}

// Indicates if the traffic matched against the rule in inbound or outbound.
func (o AdminRuleOutput) Direction() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Direction }).(pulumi.StringOutput)
}

// A unique read-only string that changes whenever the resource is updated.
func (o AdminRuleOutput) Etag() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Etag }).(pulumi.StringOutput)
}

// Whether the rule is custom or default.
// Expected value is 'Custom'.
func (o AdminRuleOutput) Kind() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Kind }).(pulumi.StringOutput)
}

// Resource name.
func (o AdminRuleOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// The priority of the rule. The value can be between 1 and 4096. The priority number must be unique for each rule in the collection. The lower the priority number, the higher the priority of the rule.
func (o AdminRuleOutput) Priority() pulumi.IntOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.IntOutput { return v.Priority }).(pulumi.IntOutput)
}

// Network protocol this rule applies to.
func (o AdminRuleOutput) Protocol() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Protocol }).(pulumi.StringOutput)
}

// The provisioning state of the resource.
func (o AdminRuleOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.ProvisioningState }).(pulumi.StringOutput)
}

// Unique identifier for this resource.
func (o AdminRuleOutput) ResourceGuid() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.ResourceGuid }).(pulumi.StringOutput)
}

// The source port ranges.
func (o AdminRuleOutput) SourcePortRanges() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringArrayOutput { return v.SourcePortRanges }).(pulumi.StringArrayOutput)
}

// The CIDR or source IP ranges.
func (o AdminRuleOutput) Sources() AddressPrefixItemResponseArrayOutput {
	return o.ApplyT(func(v *AdminRule) AddressPrefixItemResponseArrayOutput { return v.Sources }).(AddressPrefixItemResponseArrayOutput)
}

// The system metadata related to this resource.
func (o AdminRuleOutput) SystemData() SystemDataResponseOutput {
	return o.ApplyT(func(v *AdminRule) SystemDataResponseOutput { return v.SystemData }).(SystemDataResponseOutput)
}

// Resource type.
func (o AdminRuleOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *AdminRule) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(AdminRuleOutput{})
}
