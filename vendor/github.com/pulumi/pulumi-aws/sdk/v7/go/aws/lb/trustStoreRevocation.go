// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package lb

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provides a ELBv2 Trust Store Revocation for use with Application Load Balancer Listener resources.
//
// ## Example Usage
//
// ### Trust Store With Revocations
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			test, err := lb.NewTrustStore(ctx, "test", &lb.TrustStoreArgs{
//				Name:                         pulumi.String("tf-example-lb-ts"),
//				CaCertificatesBundleS3Bucket: pulumi.String("..."),
//				CaCertificatesBundleS3Key:    pulumi.String("..."),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = lb.NewTrustStoreRevocation(ctx, "test", &lb.TrustStoreRevocationArgs{
//				TrustStoreArn:       test.Arn,
//				RevocationsS3Bucket: pulumi.String("..."),
//				RevocationsS3Key:    pulumi.String("..."),
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
// Using `pulumi import`, import Trust Store Revocations using their ARN. For example:
//
// ```sh
// $ pulumi import aws:lb/trustStoreRevocation:TrustStoreRevocation example arn:aws:elasticloadbalancing:us-west-2:187416307283:truststore/my-trust-store/20cfe21448b66314,6
// ```
type TrustStoreRevocation struct {
	pulumi.CustomResourceState

	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// AWS assigned RevocationId, (number).
	RevocationId pulumi.IntOutput `pulumi:"revocationId"`
	// S3 Bucket name holding the client certificate CA bundle.
	RevocationsS3Bucket pulumi.StringOutput `pulumi:"revocationsS3Bucket"`
	// S3 object key holding the client certificate CA bundle.
	RevocationsS3Key pulumi.StringOutput `pulumi:"revocationsS3Key"`
	// Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.
	RevocationsS3ObjectVersion pulumi.StringPtrOutput `pulumi:"revocationsS3ObjectVersion"`
	// Trust Store ARN.
	TrustStoreArn pulumi.StringOutput `pulumi:"trustStoreArn"`
}

// NewTrustStoreRevocation registers a new resource with the given unique name, arguments, and options.
func NewTrustStoreRevocation(ctx *pulumi.Context,
	name string, args *TrustStoreRevocationArgs, opts ...pulumi.ResourceOption) (*TrustStoreRevocation, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.RevocationsS3Bucket == nil {
		return nil, errors.New("invalid value for required argument 'RevocationsS3Bucket'")
	}
	if args.RevocationsS3Key == nil {
		return nil, errors.New("invalid value for required argument 'RevocationsS3Key'")
	}
	if args.TrustStoreArn == nil {
		return nil, errors.New("invalid value for required argument 'TrustStoreArn'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource TrustStoreRevocation
	err := ctx.RegisterResource("aws:lb/trustStoreRevocation:TrustStoreRevocation", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetTrustStoreRevocation gets an existing TrustStoreRevocation resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetTrustStoreRevocation(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *TrustStoreRevocationState, opts ...pulumi.ResourceOption) (*TrustStoreRevocation, error) {
	var resource TrustStoreRevocation
	err := ctx.ReadResource("aws:lb/trustStoreRevocation:TrustStoreRevocation", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering TrustStoreRevocation resources.
type trustStoreRevocationState struct {
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// AWS assigned RevocationId, (number).
	RevocationId *int `pulumi:"revocationId"`
	// S3 Bucket name holding the client certificate CA bundle.
	RevocationsS3Bucket *string `pulumi:"revocationsS3Bucket"`
	// S3 object key holding the client certificate CA bundle.
	RevocationsS3Key *string `pulumi:"revocationsS3Key"`
	// Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.
	RevocationsS3ObjectVersion *string `pulumi:"revocationsS3ObjectVersion"`
	// Trust Store ARN.
	TrustStoreArn *string `pulumi:"trustStoreArn"`
}

type TrustStoreRevocationState struct {
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// AWS assigned RevocationId, (number).
	RevocationId pulumi.IntPtrInput
	// S3 Bucket name holding the client certificate CA bundle.
	RevocationsS3Bucket pulumi.StringPtrInput
	// S3 object key holding the client certificate CA bundle.
	RevocationsS3Key pulumi.StringPtrInput
	// Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.
	RevocationsS3ObjectVersion pulumi.StringPtrInput
	// Trust Store ARN.
	TrustStoreArn pulumi.StringPtrInput
}

func (TrustStoreRevocationState) ElementType() reflect.Type {
	return reflect.TypeOf((*trustStoreRevocationState)(nil)).Elem()
}

type trustStoreRevocationArgs struct {
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// S3 Bucket name holding the client certificate CA bundle.
	RevocationsS3Bucket string `pulumi:"revocationsS3Bucket"`
	// S3 object key holding the client certificate CA bundle.
	RevocationsS3Key string `pulumi:"revocationsS3Key"`
	// Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.
	RevocationsS3ObjectVersion *string `pulumi:"revocationsS3ObjectVersion"`
	// Trust Store ARN.
	TrustStoreArn string `pulumi:"trustStoreArn"`
}

// The set of arguments for constructing a TrustStoreRevocation resource.
type TrustStoreRevocationArgs struct {
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// S3 Bucket name holding the client certificate CA bundle.
	RevocationsS3Bucket pulumi.StringInput
	// S3 object key holding the client certificate CA bundle.
	RevocationsS3Key pulumi.StringInput
	// Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.
	RevocationsS3ObjectVersion pulumi.StringPtrInput
	// Trust Store ARN.
	TrustStoreArn pulumi.StringInput
}

func (TrustStoreRevocationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*trustStoreRevocationArgs)(nil)).Elem()
}

type TrustStoreRevocationInput interface {
	pulumi.Input

	ToTrustStoreRevocationOutput() TrustStoreRevocationOutput
	ToTrustStoreRevocationOutputWithContext(ctx context.Context) TrustStoreRevocationOutput
}

func (*TrustStoreRevocation) ElementType() reflect.Type {
	return reflect.TypeOf((**TrustStoreRevocation)(nil)).Elem()
}

func (i *TrustStoreRevocation) ToTrustStoreRevocationOutput() TrustStoreRevocationOutput {
	return i.ToTrustStoreRevocationOutputWithContext(context.Background())
}

func (i *TrustStoreRevocation) ToTrustStoreRevocationOutputWithContext(ctx context.Context) TrustStoreRevocationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TrustStoreRevocationOutput)
}

// TrustStoreRevocationArrayInput is an input type that accepts TrustStoreRevocationArray and TrustStoreRevocationArrayOutput values.
// You can construct a concrete instance of `TrustStoreRevocationArrayInput` via:
//
//	TrustStoreRevocationArray{ TrustStoreRevocationArgs{...} }
type TrustStoreRevocationArrayInput interface {
	pulumi.Input

	ToTrustStoreRevocationArrayOutput() TrustStoreRevocationArrayOutput
	ToTrustStoreRevocationArrayOutputWithContext(context.Context) TrustStoreRevocationArrayOutput
}

type TrustStoreRevocationArray []TrustStoreRevocationInput

func (TrustStoreRevocationArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*TrustStoreRevocation)(nil)).Elem()
}

func (i TrustStoreRevocationArray) ToTrustStoreRevocationArrayOutput() TrustStoreRevocationArrayOutput {
	return i.ToTrustStoreRevocationArrayOutputWithContext(context.Background())
}

func (i TrustStoreRevocationArray) ToTrustStoreRevocationArrayOutputWithContext(ctx context.Context) TrustStoreRevocationArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TrustStoreRevocationArrayOutput)
}

// TrustStoreRevocationMapInput is an input type that accepts TrustStoreRevocationMap and TrustStoreRevocationMapOutput values.
// You can construct a concrete instance of `TrustStoreRevocationMapInput` via:
//
//	TrustStoreRevocationMap{ "key": TrustStoreRevocationArgs{...} }
type TrustStoreRevocationMapInput interface {
	pulumi.Input

	ToTrustStoreRevocationMapOutput() TrustStoreRevocationMapOutput
	ToTrustStoreRevocationMapOutputWithContext(context.Context) TrustStoreRevocationMapOutput
}

type TrustStoreRevocationMap map[string]TrustStoreRevocationInput

func (TrustStoreRevocationMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*TrustStoreRevocation)(nil)).Elem()
}

func (i TrustStoreRevocationMap) ToTrustStoreRevocationMapOutput() TrustStoreRevocationMapOutput {
	return i.ToTrustStoreRevocationMapOutputWithContext(context.Background())
}

func (i TrustStoreRevocationMap) ToTrustStoreRevocationMapOutputWithContext(ctx context.Context) TrustStoreRevocationMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TrustStoreRevocationMapOutput)
}

type TrustStoreRevocationOutput struct{ *pulumi.OutputState }

func (TrustStoreRevocationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**TrustStoreRevocation)(nil)).Elem()
}

func (o TrustStoreRevocationOutput) ToTrustStoreRevocationOutput() TrustStoreRevocationOutput {
	return o
}

func (o TrustStoreRevocationOutput) ToTrustStoreRevocationOutputWithContext(ctx context.Context) TrustStoreRevocationOutput {
	return o
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o TrustStoreRevocationOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *TrustStoreRevocation) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// AWS assigned RevocationId, (number).
func (o TrustStoreRevocationOutput) RevocationId() pulumi.IntOutput {
	return o.ApplyT(func(v *TrustStoreRevocation) pulumi.IntOutput { return v.RevocationId }).(pulumi.IntOutput)
}

// S3 Bucket name holding the client certificate CA bundle.
func (o TrustStoreRevocationOutput) RevocationsS3Bucket() pulumi.StringOutput {
	return o.ApplyT(func(v *TrustStoreRevocation) pulumi.StringOutput { return v.RevocationsS3Bucket }).(pulumi.StringOutput)
}

// S3 object key holding the client certificate CA bundle.
func (o TrustStoreRevocationOutput) RevocationsS3Key() pulumi.StringOutput {
	return o.ApplyT(func(v *TrustStoreRevocation) pulumi.StringOutput { return v.RevocationsS3Key }).(pulumi.StringOutput)
}

// Version Id of CA bundle S3 bucket object, if versioned, defaults to latest if omitted.
func (o TrustStoreRevocationOutput) RevocationsS3ObjectVersion() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *TrustStoreRevocation) pulumi.StringPtrOutput { return v.RevocationsS3ObjectVersion }).(pulumi.StringPtrOutput)
}

// Trust Store ARN.
func (o TrustStoreRevocationOutput) TrustStoreArn() pulumi.StringOutput {
	return o.ApplyT(func(v *TrustStoreRevocation) pulumi.StringOutput { return v.TrustStoreArn }).(pulumi.StringOutput)
}

type TrustStoreRevocationArrayOutput struct{ *pulumi.OutputState }

func (TrustStoreRevocationArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*TrustStoreRevocation)(nil)).Elem()
}

func (o TrustStoreRevocationArrayOutput) ToTrustStoreRevocationArrayOutput() TrustStoreRevocationArrayOutput {
	return o
}

func (o TrustStoreRevocationArrayOutput) ToTrustStoreRevocationArrayOutputWithContext(ctx context.Context) TrustStoreRevocationArrayOutput {
	return o
}

func (o TrustStoreRevocationArrayOutput) Index(i pulumi.IntInput) TrustStoreRevocationOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *TrustStoreRevocation {
		return vs[0].([]*TrustStoreRevocation)[vs[1].(int)]
	}).(TrustStoreRevocationOutput)
}

type TrustStoreRevocationMapOutput struct{ *pulumi.OutputState }

func (TrustStoreRevocationMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*TrustStoreRevocation)(nil)).Elem()
}

func (o TrustStoreRevocationMapOutput) ToTrustStoreRevocationMapOutput() TrustStoreRevocationMapOutput {
	return o
}

func (o TrustStoreRevocationMapOutput) ToTrustStoreRevocationMapOutputWithContext(ctx context.Context) TrustStoreRevocationMapOutput {
	return o
}

func (o TrustStoreRevocationMapOutput) MapIndex(k pulumi.StringInput) TrustStoreRevocationOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *TrustStoreRevocation {
		return vs[0].(map[string]*TrustStoreRevocation)[vs[1].(string)]
	}).(TrustStoreRevocationOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*TrustStoreRevocationInput)(nil)).Elem(), &TrustStoreRevocation{})
	pulumi.RegisterInputType(reflect.TypeOf((*TrustStoreRevocationArrayInput)(nil)).Elem(), TrustStoreRevocationArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*TrustStoreRevocationMapInput)(nil)).Elem(), TrustStoreRevocationMap{})
	pulumi.RegisterOutputType(TrustStoreRevocationOutput{})
	pulumi.RegisterOutputType(TrustStoreRevocationArrayOutput{})
	pulumi.RegisterOutputType(TrustStoreRevocationMapOutput{})
}
