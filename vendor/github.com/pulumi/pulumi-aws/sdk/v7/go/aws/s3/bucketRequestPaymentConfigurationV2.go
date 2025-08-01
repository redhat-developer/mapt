// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package s3

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provides an S3 bucket request payment configuration resource. For more information, see [Requester Pays Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html).
//
// > **NOTE:** Destroying an `s3.BucketRequestPaymentConfiguration` resource resets the bucket's `payer` to the S3 default: the bucket owner.
//
// > This resource cannot be used with S3 directory buckets.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := s3.NewBucketRequestPaymentConfiguration(ctx, "example", &s3.BucketRequestPaymentConfigurationArgs{
//				Bucket: pulumi.Any(exampleAwsS3Bucket.Id),
//				Payer:  pulumi.String("Requester"),
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
// If the owner (account ID) of the source bucket differs from the account used to configure the AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// __Using `pulumi import` to import__ S3 bucket request payment configuration using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:
//
// If the owner (account ID) of the source bucket is the same account used to configure the AWS Provider, import using the `bucket`:
//
// ```sh
// $ pulumi import aws:s3/bucketRequestPaymentConfigurationV2:BucketRequestPaymentConfigurationV2 example bucket-name
// ```
// If the owner (account ID) of the source bucket differs from the account used to configure the AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// ```sh
// $ pulumi import aws:s3/bucketRequestPaymentConfigurationV2:BucketRequestPaymentConfigurationV2 example bucket-name,123456789012
// ```
//
// Deprecated: aws.s3/bucketrequestpaymentconfigurationv2.BucketRequestPaymentConfigurationV2 has been deprecated in favor of aws.s3/bucketrequestpaymentconfiguration.BucketRequestPaymentConfiguration
type BucketRequestPaymentConfigurationV2 struct {
	pulumi.CustomResourceState

	// Name of the bucket.
	Bucket pulumi.StringOutput `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrOutput `pulumi:"expectedBucketOwner"`
	// Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.
	Payer pulumi.StringOutput `pulumi:"payer"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
}

// NewBucketRequestPaymentConfigurationV2 registers a new resource with the given unique name, arguments, and options.
func NewBucketRequestPaymentConfigurationV2(ctx *pulumi.Context,
	name string, args *BucketRequestPaymentConfigurationV2Args, opts ...pulumi.ResourceOption) (*BucketRequestPaymentConfigurationV2, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Bucket == nil {
		return nil, errors.New("invalid value for required argument 'Bucket'")
	}
	if args.Payer == nil {
		return nil, errors.New("invalid value for required argument 'Payer'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("aws:s3/bucketRequestPaymentConfigurationV2:BucketRequestPaymentConfigurationV2"),
		},
	})
	opts = append(opts, aliases)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource BucketRequestPaymentConfigurationV2
	err := ctx.RegisterResource("aws:s3/bucketRequestPaymentConfigurationV2:BucketRequestPaymentConfigurationV2", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetBucketRequestPaymentConfigurationV2 gets an existing BucketRequestPaymentConfigurationV2 resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetBucketRequestPaymentConfigurationV2(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *BucketRequestPaymentConfigurationV2State, opts ...pulumi.ResourceOption) (*BucketRequestPaymentConfigurationV2, error) {
	var resource BucketRequestPaymentConfigurationV2
	err := ctx.ReadResource("aws:s3/bucketRequestPaymentConfigurationV2:BucketRequestPaymentConfigurationV2", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering BucketRequestPaymentConfigurationV2 resources.
type bucketRequestPaymentConfigurationV2State struct {
	// Name of the bucket.
	Bucket *string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.
	Payer *string `pulumi:"payer"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

type BucketRequestPaymentConfigurationV2State struct {
	// Name of the bucket.
	Bucket pulumi.StringPtrInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.
	Payer pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
}

func (BucketRequestPaymentConfigurationV2State) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketRequestPaymentConfigurationV2State)(nil)).Elem()
}

type bucketRequestPaymentConfigurationV2Args struct {
	// Name of the bucket.
	Bucket string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.
	Payer string `pulumi:"payer"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

// The set of arguments for constructing a BucketRequestPaymentConfigurationV2 resource.
type BucketRequestPaymentConfigurationV2Args struct {
	// Name of the bucket.
	Bucket pulumi.StringInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.
	Payer pulumi.StringInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
}

func (BucketRequestPaymentConfigurationV2Args) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketRequestPaymentConfigurationV2Args)(nil)).Elem()
}

type BucketRequestPaymentConfigurationV2Input interface {
	pulumi.Input

	ToBucketRequestPaymentConfigurationV2Output() BucketRequestPaymentConfigurationV2Output
	ToBucketRequestPaymentConfigurationV2OutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2Output
}

func (*BucketRequestPaymentConfigurationV2) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketRequestPaymentConfigurationV2)(nil)).Elem()
}

func (i *BucketRequestPaymentConfigurationV2) ToBucketRequestPaymentConfigurationV2Output() BucketRequestPaymentConfigurationV2Output {
	return i.ToBucketRequestPaymentConfigurationV2OutputWithContext(context.Background())
}

func (i *BucketRequestPaymentConfigurationV2) ToBucketRequestPaymentConfigurationV2OutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2Output {
	return pulumi.ToOutputWithContext(ctx, i).(BucketRequestPaymentConfigurationV2Output)
}

// BucketRequestPaymentConfigurationV2ArrayInput is an input type that accepts BucketRequestPaymentConfigurationV2Array and BucketRequestPaymentConfigurationV2ArrayOutput values.
// You can construct a concrete instance of `BucketRequestPaymentConfigurationV2ArrayInput` via:
//
//	BucketRequestPaymentConfigurationV2Array{ BucketRequestPaymentConfigurationV2Args{...} }
type BucketRequestPaymentConfigurationV2ArrayInput interface {
	pulumi.Input

	ToBucketRequestPaymentConfigurationV2ArrayOutput() BucketRequestPaymentConfigurationV2ArrayOutput
	ToBucketRequestPaymentConfigurationV2ArrayOutputWithContext(context.Context) BucketRequestPaymentConfigurationV2ArrayOutput
}

type BucketRequestPaymentConfigurationV2Array []BucketRequestPaymentConfigurationV2Input

func (BucketRequestPaymentConfigurationV2Array) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketRequestPaymentConfigurationV2)(nil)).Elem()
}

func (i BucketRequestPaymentConfigurationV2Array) ToBucketRequestPaymentConfigurationV2ArrayOutput() BucketRequestPaymentConfigurationV2ArrayOutput {
	return i.ToBucketRequestPaymentConfigurationV2ArrayOutputWithContext(context.Background())
}

func (i BucketRequestPaymentConfigurationV2Array) ToBucketRequestPaymentConfigurationV2ArrayOutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2ArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketRequestPaymentConfigurationV2ArrayOutput)
}

// BucketRequestPaymentConfigurationV2MapInput is an input type that accepts BucketRequestPaymentConfigurationV2Map and BucketRequestPaymentConfigurationV2MapOutput values.
// You can construct a concrete instance of `BucketRequestPaymentConfigurationV2MapInput` via:
//
//	BucketRequestPaymentConfigurationV2Map{ "key": BucketRequestPaymentConfigurationV2Args{...} }
type BucketRequestPaymentConfigurationV2MapInput interface {
	pulumi.Input

	ToBucketRequestPaymentConfigurationV2MapOutput() BucketRequestPaymentConfigurationV2MapOutput
	ToBucketRequestPaymentConfigurationV2MapOutputWithContext(context.Context) BucketRequestPaymentConfigurationV2MapOutput
}

type BucketRequestPaymentConfigurationV2Map map[string]BucketRequestPaymentConfigurationV2Input

func (BucketRequestPaymentConfigurationV2Map) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketRequestPaymentConfigurationV2)(nil)).Elem()
}

func (i BucketRequestPaymentConfigurationV2Map) ToBucketRequestPaymentConfigurationV2MapOutput() BucketRequestPaymentConfigurationV2MapOutput {
	return i.ToBucketRequestPaymentConfigurationV2MapOutputWithContext(context.Background())
}

func (i BucketRequestPaymentConfigurationV2Map) ToBucketRequestPaymentConfigurationV2MapOutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2MapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketRequestPaymentConfigurationV2MapOutput)
}

type BucketRequestPaymentConfigurationV2Output struct{ *pulumi.OutputState }

func (BucketRequestPaymentConfigurationV2Output) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketRequestPaymentConfigurationV2)(nil)).Elem()
}

func (o BucketRequestPaymentConfigurationV2Output) ToBucketRequestPaymentConfigurationV2Output() BucketRequestPaymentConfigurationV2Output {
	return o
}

func (o BucketRequestPaymentConfigurationV2Output) ToBucketRequestPaymentConfigurationV2OutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2Output {
	return o
}

// Name of the bucket.
func (o BucketRequestPaymentConfigurationV2Output) Bucket() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketRequestPaymentConfigurationV2) pulumi.StringOutput { return v.Bucket }).(pulumi.StringOutput)
}

// Account ID of the expected bucket owner.
func (o BucketRequestPaymentConfigurationV2Output) ExpectedBucketOwner() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *BucketRequestPaymentConfigurationV2) pulumi.StringPtrOutput { return v.ExpectedBucketOwner }).(pulumi.StringPtrOutput)
}

// Specifies who pays for the download and request fees. Valid values: `BucketOwner`, `Requester`.
func (o BucketRequestPaymentConfigurationV2Output) Payer() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketRequestPaymentConfigurationV2) pulumi.StringOutput { return v.Payer }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o BucketRequestPaymentConfigurationV2Output) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketRequestPaymentConfigurationV2) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

type BucketRequestPaymentConfigurationV2ArrayOutput struct{ *pulumi.OutputState }

func (BucketRequestPaymentConfigurationV2ArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketRequestPaymentConfigurationV2)(nil)).Elem()
}

func (o BucketRequestPaymentConfigurationV2ArrayOutput) ToBucketRequestPaymentConfigurationV2ArrayOutput() BucketRequestPaymentConfigurationV2ArrayOutput {
	return o
}

func (o BucketRequestPaymentConfigurationV2ArrayOutput) ToBucketRequestPaymentConfigurationV2ArrayOutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2ArrayOutput {
	return o
}

func (o BucketRequestPaymentConfigurationV2ArrayOutput) Index(i pulumi.IntInput) BucketRequestPaymentConfigurationV2Output {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *BucketRequestPaymentConfigurationV2 {
		return vs[0].([]*BucketRequestPaymentConfigurationV2)[vs[1].(int)]
	}).(BucketRequestPaymentConfigurationV2Output)
}

type BucketRequestPaymentConfigurationV2MapOutput struct{ *pulumi.OutputState }

func (BucketRequestPaymentConfigurationV2MapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketRequestPaymentConfigurationV2)(nil)).Elem()
}

func (o BucketRequestPaymentConfigurationV2MapOutput) ToBucketRequestPaymentConfigurationV2MapOutput() BucketRequestPaymentConfigurationV2MapOutput {
	return o
}

func (o BucketRequestPaymentConfigurationV2MapOutput) ToBucketRequestPaymentConfigurationV2MapOutputWithContext(ctx context.Context) BucketRequestPaymentConfigurationV2MapOutput {
	return o
}

func (o BucketRequestPaymentConfigurationV2MapOutput) MapIndex(k pulumi.StringInput) BucketRequestPaymentConfigurationV2Output {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *BucketRequestPaymentConfigurationV2 {
		return vs[0].(map[string]*BucketRequestPaymentConfigurationV2)[vs[1].(string)]
	}).(BucketRequestPaymentConfigurationV2Output)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*BucketRequestPaymentConfigurationV2Input)(nil)).Elem(), &BucketRequestPaymentConfigurationV2{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketRequestPaymentConfigurationV2ArrayInput)(nil)).Elem(), BucketRequestPaymentConfigurationV2Array{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketRequestPaymentConfigurationV2MapInput)(nil)).Elem(), BucketRequestPaymentConfigurationV2Map{})
	pulumi.RegisterOutputType(BucketRequestPaymentConfigurationV2Output{})
	pulumi.RegisterOutputType(BucketRequestPaymentConfigurationV2ArrayOutput{})
	pulumi.RegisterOutputType(BucketRequestPaymentConfigurationV2MapOutput{})
}
