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

// Provides a resource to allow a principal to discover a VPC endpoint service.
//
// > **NOTE on VPC Endpoint Services and VPC Endpoint Service Allowed Principals:** This provider provides
// both a standalone VPC Endpoint Service Allowed Principal resource
// and a VPC Endpoint Service resource with an `allowedPrincipals` attribute. Do not use the same principal ARN in both
// a VPC Endpoint Service resource and a VPC Endpoint Service Allowed Principal resource. Doing so will cause a conflict
// and will overwrite the association.
//
// ## Example Usage
//
// Basic usage:
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			current, err := aws.GetCallerIdentity(ctx, &aws.GetCallerIdentityArgs{}, nil)
//			if err != nil {
//				return err
//			}
//			_, err = ec2.NewVpcEndpointServiceAllowedPrinciple(ctx, "allow_me_to_foo", &ec2.VpcEndpointServiceAllowedPrincipleArgs{
//				VpcEndpointServiceId: pulumi.Any(foo.Id),
//				PrincipalArn:         pulumi.String(current.Arn),
//			})
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
type VpcEndpointServiceAllowedPrinciple struct {
	pulumi.CustomResourceState

	// The ARN of the principal to allow permissions.
	PrincipalArn pulumi.StringOutput `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// The ID of the VPC endpoint service to allow permission.
	VpcEndpointServiceId pulumi.StringOutput `pulumi:"vpcEndpointServiceId"`
}

// NewVpcEndpointServiceAllowedPrinciple registers a new resource with the given unique name, arguments, and options.
func NewVpcEndpointServiceAllowedPrinciple(ctx *pulumi.Context,
	name string, args *VpcEndpointServiceAllowedPrincipleArgs, opts ...pulumi.ResourceOption) (*VpcEndpointServiceAllowedPrinciple, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.PrincipalArn == nil {
		return nil, errors.New("invalid value for required argument 'PrincipalArn'")
	}
	if args.VpcEndpointServiceId == nil {
		return nil, errors.New("invalid value for required argument 'VpcEndpointServiceId'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource VpcEndpointServiceAllowedPrinciple
	err := ctx.RegisterResource("aws:ec2/vpcEndpointServiceAllowedPrinciple:VpcEndpointServiceAllowedPrinciple", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetVpcEndpointServiceAllowedPrinciple gets an existing VpcEndpointServiceAllowedPrinciple resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetVpcEndpointServiceAllowedPrinciple(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *VpcEndpointServiceAllowedPrincipleState, opts ...pulumi.ResourceOption) (*VpcEndpointServiceAllowedPrinciple, error) {
	var resource VpcEndpointServiceAllowedPrinciple
	err := ctx.ReadResource("aws:ec2/vpcEndpointServiceAllowedPrinciple:VpcEndpointServiceAllowedPrinciple", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering VpcEndpointServiceAllowedPrinciple resources.
type vpcEndpointServiceAllowedPrincipleState struct {
	// The ARN of the principal to allow permissions.
	PrincipalArn *string `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The ID of the VPC endpoint service to allow permission.
	VpcEndpointServiceId *string `pulumi:"vpcEndpointServiceId"`
}

type VpcEndpointServiceAllowedPrincipleState struct {
	// The ARN of the principal to allow permissions.
	PrincipalArn pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The ID of the VPC endpoint service to allow permission.
	VpcEndpointServiceId pulumi.StringPtrInput
}

func (VpcEndpointServiceAllowedPrincipleState) ElementType() reflect.Type {
	return reflect.TypeOf((*vpcEndpointServiceAllowedPrincipleState)(nil)).Elem()
}

type vpcEndpointServiceAllowedPrincipleArgs struct {
	// The ARN of the principal to allow permissions.
	PrincipalArn string `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The ID of the VPC endpoint service to allow permission.
	VpcEndpointServiceId string `pulumi:"vpcEndpointServiceId"`
}

// The set of arguments for constructing a VpcEndpointServiceAllowedPrinciple resource.
type VpcEndpointServiceAllowedPrincipleArgs struct {
	// The ARN of the principal to allow permissions.
	PrincipalArn pulumi.StringInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The ID of the VPC endpoint service to allow permission.
	VpcEndpointServiceId pulumi.StringInput
}

func (VpcEndpointServiceAllowedPrincipleArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*vpcEndpointServiceAllowedPrincipleArgs)(nil)).Elem()
}

type VpcEndpointServiceAllowedPrincipleInput interface {
	pulumi.Input

	ToVpcEndpointServiceAllowedPrincipleOutput() VpcEndpointServiceAllowedPrincipleOutput
	ToVpcEndpointServiceAllowedPrincipleOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleOutput
}

func (*VpcEndpointServiceAllowedPrinciple) ElementType() reflect.Type {
	return reflect.TypeOf((**VpcEndpointServiceAllowedPrinciple)(nil)).Elem()
}

func (i *VpcEndpointServiceAllowedPrinciple) ToVpcEndpointServiceAllowedPrincipleOutput() VpcEndpointServiceAllowedPrincipleOutput {
	return i.ToVpcEndpointServiceAllowedPrincipleOutputWithContext(context.Background())
}

func (i *VpcEndpointServiceAllowedPrinciple) ToVpcEndpointServiceAllowedPrincipleOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleOutput {
	return pulumi.ToOutputWithContext(ctx, i).(VpcEndpointServiceAllowedPrincipleOutput)
}

// VpcEndpointServiceAllowedPrincipleArrayInput is an input type that accepts VpcEndpointServiceAllowedPrincipleArray and VpcEndpointServiceAllowedPrincipleArrayOutput values.
// You can construct a concrete instance of `VpcEndpointServiceAllowedPrincipleArrayInput` via:
//
//	VpcEndpointServiceAllowedPrincipleArray{ VpcEndpointServiceAllowedPrincipleArgs{...} }
type VpcEndpointServiceAllowedPrincipleArrayInput interface {
	pulumi.Input

	ToVpcEndpointServiceAllowedPrincipleArrayOutput() VpcEndpointServiceAllowedPrincipleArrayOutput
	ToVpcEndpointServiceAllowedPrincipleArrayOutputWithContext(context.Context) VpcEndpointServiceAllowedPrincipleArrayOutput
}

type VpcEndpointServiceAllowedPrincipleArray []VpcEndpointServiceAllowedPrincipleInput

func (VpcEndpointServiceAllowedPrincipleArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*VpcEndpointServiceAllowedPrinciple)(nil)).Elem()
}

func (i VpcEndpointServiceAllowedPrincipleArray) ToVpcEndpointServiceAllowedPrincipleArrayOutput() VpcEndpointServiceAllowedPrincipleArrayOutput {
	return i.ToVpcEndpointServiceAllowedPrincipleArrayOutputWithContext(context.Background())
}

func (i VpcEndpointServiceAllowedPrincipleArray) ToVpcEndpointServiceAllowedPrincipleArrayOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(VpcEndpointServiceAllowedPrincipleArrayOutput)
}

// VpcEndpointServiceAllowedPrincipleMapInput is an input type that accepts VpcEndpointServiceAllowedPrincipleMap and VpcEndpointServiceAllowedPrincipleMapOutput values.
// You can construct a concrete instance of `VpcEndpointServiceAllowedPrincipleMapInput` via:
//
//	VpcEndpointServiceAllowedPrincipleMap{ "key": VpcEndpointServiceAllowedPrincipleArgs{...} }
type VpcEndpointServiceAllowedPrincipleMapInput interface {
	pulumi.Input

	ToVpcEndpointServiceAllowedPrincipleMapOutput() VpcEndpointServiceAllowedPrincipleMapOutput
	ToVpcEndpointServiceAllowedPrincipleMapOutputWithContext(context.Context) VpcEndpointServiceAllowedPrincipleMapOutput
}

type VpcEndpointServiceAllowedPrincipleMap map[string]VpcEndpointServiceAllowedPrincipleInput

func (VpcEndpointServiceAllowedPrincipleMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*VpcEndpointServiceAllowedPrinciple)(nil)).Elem()
}

func (i VpcEndpointServiceAllowedPrincipleMap) ToVpcEndpointServiceAllowedPrincipleMapOutput() VpcEndpointServiceAllowedPrincipleMapOutput {
	return i.ToVpcEndpointServiceAllowedPrincipleMapOutputWithContext(context.Background())
}

func (i VpcEndpointServiceAllowedPrincipleMap) ToVpcEndpointServiceAllowedPrincipleMapOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(VpcEndpointServiceAllowedPrincipleMapOutput)
}

type VpcEndpointServiceAllowedPrincipleOutput struct{ *pulumi.OutputState }

func (VpcEndpointServiceAllowedPrincipleOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**VpcEndpointServiceAllowedPrinciple)(nil)).Elem()
}

func (o VpcEndpointServiceAllowedPrincipleOutput) ToVpcEndpointServiceAllowedPrincipleOutput() VpcEndpointServiceAllowedPrincipleOutput {
	return o
}

func (o VpcEndpointServiceAllowedPrincipleOutput) ToVpcEndpointServiceAllowedPrincipleOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleOutput {
	return o
}

// The ARN of the principal to allow permissions.
func (o VpcEndpointServiceAllowedPrincipleOutput) PrincipalArn() pulumi.StringOutput {
	return o.ApplyT(func(v *VpcEndpointServiceAllowedPrinciple) pulumi.StringOutput { return v.PrincipalArn }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o VpcEndpointServiceAllowedPrincipleOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *VpcEndpointServiceAllowedPrinciple) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// The ID of the VPC endpoint service to allow permission.
func (o VpcEndpointServiceAllowedPrincipleOutput) VpcEndpointServiceId() pulumi.StringOutput {
	return o.ApplyT(func(v *VpcEndpointServiceAllowedPrinciple) pulumi.StringOutput { return v.VpcEndpointServiceId }).(pulumi.StringOutput)
}

type VpcEndpointServiceAllowedPrincipleArrayOutput struct{ *pulumi.OutputState }

func (VpcEndpointServiceAllowedPrincipleArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*VpcEndpointServiceAllowedPrinciple)(nil)).Elem()
}

func (o VpcEndpointServiceAllowedPrincipleArrayOutput) ToVpcEndpointServiceAllowedPrincipleArrayOutput() VpcEndpointServiceAllowedPrincipleArrayOutput {
	return o
}

func (o VpcEndpointServiceAllowedPrincipleArrayOutput) ToVpcEndpointServiceAllowedPrincipleArrayOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleArrayOutput {
	return o
}

func (o VpcEndpointServiceAllowedPrincipleArrayOutput) Index(i pulumi.IntInput) VpcEndpointServiceAllowedPrincipleOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *VpcEndpointServiceAllowedPrinciple {
		return vs[0].([]*VpcEndpointServiceAllowedPrinciple)[vs[1].(int)]
	}).(VpcEndpointServiceAllowedPrincipleOutput)
}

type VpcEndpointServiceAllowedPrincipleMapOutput struct{ *pulumi.OutputState }

func (VpcEndpointServiceAllowedPrincipleMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*VpcEndpointServiceAllowedPrinciple)(nil)).Elem()
}

func (o VpcEndpointServiceAllowedPrincipleMapOutput) ToVpcEndpointServiceAllowedPrincipleMapOutput() VpcEndpointServiceAllowedPrincipleMapOutput {
	return o
}

func (o VpcEndpointServiceAllowedPrincipleMapOutput) ToVpcEndpointServiceAllowedPrincipleMapOutputWithContext(ctx context.Context) VpcEndpointServiceAllowedPrincipleMapOutput {
	return o
}

func (o VpcEndpointServiceAllowedPrincipleMapOutput) MapIndex(k pulumi.StringInput) VpcEndpointServiceAllowedPrincipleOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *VpcEndpointServiceAllowedPrinciple {
		return vs[0].(map[string]*VpcEndpointServiceAllowedPrinciple)[vs[1].(string)]
	}).(VpcEndpointServiceAllowedPrincipleOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*VpcEndpointServiceAllowedPrincipleInput)(nil)).Elem(), &VpcEndpointServiceAllowedPrinciple{})
	pulumi.RegisterInputType(reflect.TypeOf((*VpcEndpointServiceAllowedPrincipleArrayInput)(nil)).Elem(), VpcEndpointServiceAllowedPrincipleArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*VpcEndpointServiceAllowedPrincipleMapInput)(nil)).Elem(), VpcEndpointServiceAllowedPrincipleMap{})
	pulumi.RegisterOutputType(VpcEndpointServiceAllowedPrincipleOutput{})
	pulumi.RegisterOutputType(VpcEndpointServiceAllowedPrincipleArrayOutput{})
	pulumi.RegisterOutputType(VpcEndpointServiceAllowedPrincipleMapOutput{})
}
