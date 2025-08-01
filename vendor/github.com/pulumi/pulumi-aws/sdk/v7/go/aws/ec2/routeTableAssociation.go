// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package ec2

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provides a resource to create an association between a route table and a subnet or a route table and an
// internet gateway or virtual private gateway.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := ec2.NewRouteTableAssociation(ctx, "a", &ec2.RouteTableAssociationArgs{
//				SubnetId:     pulumi.Any(foo.Id),
//				RouteTableId: pulumi.Any(bar.Id),
//			})
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := ec2.NewRouteTableAssociation(ctx, "b", &ec2.RouteTableAssociationArgs{
//				GatewayId:    pulumi.Any(foo.Id),
//				RouteTableId: pulumi.Any(bar.Id),
//			})
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
//
// ## Import
//
// With EC2 Internet Gateways:
//
// __Using `pulumi import` to import__ EC2 Route Table Associations using the associated resource ID and Route Table ID separated by a forward slash (`/`). For example:
//
// With EC2 Subnets:
//
// ```sh
// $ pulumi import aws:ec2/routeTableAssociation:RouteTableAssociation assoc subnet-6777656e646f6c796e/rtb-656c65616e6f72
// ```
// With EC2 Internet Gateways:
//
// ```sh
// $ pulumi import aws:ec2/routeTableAssociation:RouteTableAssociation assoc igw-01b3a60780f8d034a/rtb-656c65616e6f72
// ```
type RouteTableAssociation struct {
	pulumi.CustomResourceState

	// The gateway ID to create an association. Conflicts with `subnetId`.
	GatewayId pulumi.StringPtrOutput `pulumi:"gatewayId"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// The ID of the routing table to associate with.
	//
	// > **NOTE:** Please note that one of either `subnetId` or `gatewayId` is required.
	RouteTableId pulumi.StringOutput `pulumi:"routeTableId"`
	// The subnet ID to create an association. Conflicts with `gatewayId`.
	SubnetId pulumi.StringPtrOutput `pulumi:"subnetId"`
}

// NewRouteTableAssociation registers a new resource with the given unique name, arguments, and options.
func NewRouteTableAssociation(ctx *pulumi.Context,
	name string, args *RouteTableAssociationArgs, opts ...pulumi.ResourceOption) (*RouteTableAssociation, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.RouteTableId == nil {
		return nil, errors.New("invalid value for required argument 'RouteTableId'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource RouteTableAssociation
	err := ctx.RegisterResource("aws:ec2/routeTableAssociation:RouteTableAssociation", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetRouteTableAssociation gets an existing RouteTableAssociation resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetRouteTableAssociation(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *RouteTableAssociationState, opts ...pulumi.ResourceOption) (*RouteTableAssociation, error) {
	var resource RouteTableAssociation
	err := ctx.ReadResource("aws:ec2/routeTableAssociation:RouteTableAssociation", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering RouteTableAssociation resources.
type routeTableAssociationState struct {
	// The gateway ID to create an association. Conflicts with `subnetId`.
	GatewayId *string `pulumi:"gatewayId"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The ID of the routing table to associate with.
	//
	// > **NOTE:** Please note that one of either `subnetId` or `gatewayId` is required.
	RouteTableId *string `pulumi:"routeTableId"`
	// The subnet ID to create an association. Conflicts with `gatewayId`.
	SubnetId *string `pulumi:"subnetId"`
}

type RouteTableAssociationState struct {
	// The gateway ID to create an association. Conflicts with `subnetId`.
	GatewayId pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The ID of the routing table to associate with.
	//
	// > **NOTE:** Please note that one of either `subnetId` or `gatewayId` is required.
	RouteTableId pulumi.StringPtrInput
	// The subnet ID to create an association. Conflicts with `gatewayId`.
	SubnetId pulumi.StringPtrInput
}

func (RouteTableAssociationState) ElementType() reflect.Type {
	return reflect.TypeOf((*routeTableAssociationState)(nil)).Elem()
}

type routeTableAssociationArgs struct {
	// The gateway ID to create an association. Conflicts with `subnetId`.
	GatewayId *string `pulumi:"gatewayId"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The ID of the routing table to associate with.
	//
	// > **NOTE:** Please note that one of either `subnetId` or `gatewayId` is required.
	RouteTableId string `pulumi:"routeTableId"`
	// The subnet ID to create an association. Conflicts with `gatewayId`.
	SubnetId *string `pulumi:"subnetId"`
}

// The set of arguments for constructing a RouteTableAssociation resource.
type RouteTableAssociationArgs struct {
	// The gateway ID to create an association. Conflicts with `subnetId`.
	GatewayId pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The ID of the routing table to associate with.
	//
	// > **NOTE:** Please note that one of either `subnetId` or `gatewayId` is required.
	RouteTableId pulumi.StringInput
	// The subnet ID to create an association. Conflicts with `gatewayId`.
	SubnetId pulumi.StringPtrInput
}

func (RouteTableAssociationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*routeTableAssociationArgs)(nil)).Elem()
}

type RouteTableAssociationInput interface {
	pulumi.Input

	ToRouteTableAssociationOutput() RouteTableAssociationOutput
	ToRouteTableAssociationOutputWithContext(ctx context.Context) RouteTableAssociationOutput
}

func (*RouteTableAssociation) ElementType() reflect.Type {
	return reflect.TypeOf((**RouteTableAssociation)(nil)).Elem()
}

func (i *RouteTableAssociation) ToRouteTableAssociationOutput() RouteTableAssociationOutput {
	return i.ToRouteTableAssociationOutputWithContext(context.Background())
}

func (i *RouteTableAssociation) ToRouteTableAssociationOutputWithContext(ctx context.Context) RouteTableAssociationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(RouteTableAssociationOutput)
}

// RouteTableAssociationArrayInput is an input type that accepts RouteTableAssociationArray and RouteTableAssociationArrayOutput values.
// You can construct a concrete instance of `RouteTableAssociationArrayInput` via:
//
//	RouteTableAssociationArray{ RouteTableAssociationArgs{...} }
type RouteTableAssociationArrayInput interface {
	pulumi.Input

	ToRouteTableAssociationArrayOutput() RouteTableAssociationArrayOutput
	ToRouteTableAssociationArrayOutputWithContext(context.Context) RouteTableAssociationArrayOutput
}

type RouteTableAssociationArray []RouteTableAssociationInput

func (RouteTableAssociationArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*RouteTableAssociation)(nil)).Elem()
}

func (i RouteTableAssociationArray) ToRouteTableAssociationArrayOutput() RouteTableAssociationArrayOutput {
	return i.ToRouteTableAssociationArrayOutputWithContext(context.Background())
}

func (i RouteTableAssociationArray) ToRouteTableAssociationArrayOutputWithContext(ctx context.Context) RouteTableAssociationArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(RouteTableAssociationArrayOutput)
}

// RouteTableAssociationMapInput is an input type that accepts RouteTableAssociationMap and RouteTableAssociationMapOutput values.
// You can construct a concrete instance of `RouteTableAssociationMapInput` via:
//
//	RouteTableAssociationMap{ "key": RouteTableAssociationArgs{...} }
type RouteTableAssociationMapInput interface {
	pulumi.Input

	ToRouteTableAssociationMapOutput() RouteTableAssociationMapOutput
	ToRouteTableAssociationMapOutputWithContext(context.Context) RouteTableAssociationMapOutput
}

type RouteTableAssociationMap map[string]RouteTableAssociationInput

func (RouteTableAssociationMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*RouteTableAssociation)(nil)).Elem()
}

func (i RouteTableAssociationMap) ToRouteTableAssociationMapOutput() RouteTableAssociationMapOutput {
	return i.ToRouteTableAssociationMapOutputWithContext(context.Background())
}

func (i RouteTableAssociationMap) ToRouteTableAssociationMapOutputWithContext(ctx context.Context) RouteTableAssociationMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(RouteTableAssociationMapOutput)
}

type RouteTableAssociationOutput struct{ *pulumi.OutputState }

func (RouteTableAssociationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**RouteTableAssociation)(nil)).Elem()
}

func (o RouteTableAssociationOutput) ToRouteTableAssociationOutput() RouteTableAssociationOutput {
	return o
}

func (o RouteTableAssociationOutput) ToRouteTableAssociationOutputWithContext(ctx context.Context) RouteTableAssociationOutput {
	return o
}

// The gateway ID to create an association. Conflicts with `subnetId`.
func (o RouteTableAssociationOutput) GatewayId() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *RouteTableAssociation) pulumi.StringPtrOutput { return v.GatewayId }).(pulumi.StringPtrOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o RouteTableAssociationOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *RouteTableAssociation) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// The ID of the routing table to associate with.
//
// > **NOTE:** Please note that one of either `subnetId` or `gatewayId` is required.
func (o RouteTableAssociationOutput) RouteTableId() pulumi.StringOutput {
	return o.ApplyT(func(v *RouteTableAssociation) pulumi.StringOutput { return v.RouteTableId }).(pulumi.StringOutput)
}

// The subnet ID to create an association. Conflicts with `gatewayId`.
func (o RouteTableAssociationOutput) SubnetId() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *RouteTableAssociation) pulumi.StringPtrOutput { return v.SubnetId }).(pulumi.StringPtrOutput)
}

type RouteTableAssociationArrayOutput struct{ *pulumi.OutputState }

func (RouteTableAssociationArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*RouteTableAssociation)(nil)).Elem()
}

func (o RouteTableAssociationArrayOutput) ToRouteTableAssociationArrayOutput() RouteTableAssociationArrayOutput {
	return o
}

func (o RouteTableAssociationArrayOutput) ToRouteTableAssociationArrayOutputWithContext(ctx context.Context) RouteTableAssociationArrayOutput {
	return o
}

func (o RouteTableAssociationArrayOutput) Index(i pulumi.IntInput) RouteTableAssociationOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *RouteTableAssociation {
		return vs[0].([]*RouteTableAssociation)[vs[1].(int)]
	}).(RouteTableAssociationOutput)
}

type RouteTableAssociationMapOutput struct{ *pulumi.OutputState }

func (RouteTableAssociationMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*RouteTableAssociation)(nil)).Elem()
}

func (o RouteTableAssociationMapOutput) ToRouteTableAssociationMapOutput() RouteTableAssociationMapOutput {
	return o
}

func (o RouteTableAssociationMapOutput) ToRouteTableAssociationMapOutputWithContext(ctx context.Context) RouteTableAssociationMapOutput {
	return o
}

func (o RouteTableAssociationMapOutput) MapIndex(k pulumi.StringInput) RouteTableAssociationOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *RouteTableAssociation {
		return vs[0].(map[string]*RouteTableAssociation)[vs[1].(string)]
	}).(RouteTableAssociationOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*RouteTableAssociationInput)(nil)).Elem(), &RouteTableAssociation{})
	pulumi.RegisterInputType(reflect.TypeOf((*RouteTableAssociationArrayInput)(nil)).Elem(), RouteTableAssociationArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*RouteTableAssociationMapInput)(nil)).Elem(), RouteTableAssociationMap{})
	pulumi.RegisterOutputType(RouteTableAssociationOutput{})
	pulumi.RegisterOutputType(RouteTableAssociationArrayOutput{})
	pulumi.RegisterOutputType(RouteTableAssociationMapOutput{})
}
