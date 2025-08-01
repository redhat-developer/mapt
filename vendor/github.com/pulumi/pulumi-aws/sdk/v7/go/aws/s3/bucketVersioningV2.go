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

// Provides a resource for controlling versioning on an S3 bucket.
// Deleting this resource will either suspend versioning on the associated S3 bucket or
// simply remove the resource from state if the associated S3 bucket is unversioned.
//
// For more information, see [How S3 versioning works](https://docs.aws.amazon.com/AmazonS3/latest/userguide/manage-versioning-examples.html).
//
// > **NOTE:** If you are enabling versioning on the bucket for the first time, AWS recommends that you wait for 15 minutes after enabling versioning before issuing write operations (PUT or DELETE) on objects in the bucket.
//
// > This resource cannot be used with S3 directory buckets.
//
// ## Example Usage
//
// ### With Versioning Enabled
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
//			example, err := s3.NewBucket(ctx, "example", &s3.BucketArgs{
//				Bucket: pulumi.String("example-bucket"),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketAcl(ctx, "example", &s3.BucketAclArgs{
//				Bucket: example.ID(),
//				Acl:    pulumi.String("private"),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketVersioning(ctx, "versioning_example", &s3.BucketVersioningArgs{
//				Bucket: example.ID(),
//				VersioningConfiguration: &s3.BucketVersioningVersioningConfigurationArgs{
//					Status: pulumi.String("Enabled"),
//				},
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
// ### With Versioning Disabled
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
//			example, err := s3.NewBucket(ctx, "example", &s3.BucketArgs{
//				Bucket: pulumi.String("example-bucket"),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketAcl(ctx, "example", &s3.BucketAclArgs{
//				Bucket: example.ID(),
//				Acl:    pulumi.String("private"),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketVersioning(ctx, "versioning_example", &s3.BucketVersioningArgs{
//				Bucket: example.ID(),
//				VersioningConfiguration: &s3.BucketVersioningVersioningConfigurationArgs{
//					Status: pulumi.String("Disabled"),
//				},
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
// ### Object Dependency On Versioning
//
// When you create an object whose `versionId` you need and an `s3.BucketVersioning` resource in the same configuration, you are more likely to have success by ensuring the `s3Object` depends either implicitly (see below) or explicitly (i.e., using `dependsOn = [aws_s3_bucket_versioning.example]`) on the `s3.BucketVersioning` resource.
//
// > **NOTE:** For critical and/or production S3 objects, do not create a bucket, enable versioning, and create an object in the bucket within the same configuration. Doing so will not allow the AWS-recommended 15 minutes between enabling versioning and writing to the bucket.
//
// This example shows the `aws_s3_object.example` depending implicitly on the versioning resource through the reference to `aws_s3_bucket_versioning.example.bucket` to define `bucket`:
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
//			example, err := s3.NewBucket(ctx, "example", &s3.BucketArgs{
//				Bucket: pulumi.String("yotto"),
//			})
//			if err != nil {
//				return err
//			}
//			exampleBucketVersioning, err := s3.NewBucketVersioning(ctx, "example", &s3.BucketVersioningArgs{
//				Bucket: example.ID(),
//				VersioningConfiguration: &s3.BucketVersioningVersioningConfigurationArgs{
//					Status: pulumi.String("Enabled"),
//				},
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketObjectv2(ctx, "example", &s3.BucketObjectv2Args{
//				Bucket: exampleBucketVersioning.ID(),
//				Key:    pulumi.String("droeloe"),
//				Source: pulumi.NewFileAsset("example.txt"),
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
// __Using `pulumi import` to import__ S3 bucket versioning using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:
//
// If the owner (account ID) of the source bucket is the same account used to configure the AWS Provider, import using the `bucket`:
//
// ```sh
// $ pulumi import aws:s3/bucketVersioningV2:BucketVersioningV2 example bucket-name
// ```
// If the owner (account ID) of the source bucket differs from the account used to configure the AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// ```sh
// $ pulumi import aws:s3/bucketVersioningV2:BucketVersioningV2 example bucket-name,123456789012
// ```
//
// Deprecated: aws.s3/bucketversioningv2.BucketVersioningV2 has been deprecated in favor of aws.s3/bucketversioning.BucketVersioning
type BucketVersioningV2 struct {
	pulumi.CustomResourceState

	// Name of the S3 bucket.
	Bucket pulumi.StringOutput `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrOutput `pulumi:"expectedBucketOwner"`
	// Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.
	Mfa pulumi.StringPtrOutput `pulumi:"mfa"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// Configuration block for the versioning parameters. See below.
	VersioningConfiguration BucketVersioningV2VersioningConfigurationOutput `pulumi:"versioningConfiguration"`
}

// NewBucketVersioningV2 registers a new resource with the given unique name, arguments, and options.
func NewBucketVersioningV2(ctx *pulumi.Context,
	name string, args *BucketVersioningV2Args, opts ...pulumi.ResourceOption) (*BucketVersioningV2, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Bucket == nil {
		return nil, errors.New("invalid value for required argument 'Bucket'")
	}
	if args.VersioningConfiguration == nil {
		return nil, errors.New("invalid value for required argument 'VersioningConfiguration'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("aws:s3/bucketVersioningV2:BucketVersioningV2"),
		},
	})
	opts = append(opts, aliases)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource BucketVersioningV2
	err := ctx.RegisterResource("aws:s3/bucketVersioningV2:BucketVersioningV2", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetBucketVersioningV2 gets an existing BucketVersioningV2 resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetBucketVersioningV2(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *BucketVersioningV2State, opts ...pulumi.ResourceOption) (*BucketVersioningV2, error) {
	var resource BucketVersioningV2
	err := ctx.ReadResource("aws:s3/bucketVersioningV2:BucketVersioningV2", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering BucketVersioningV2 resources.
type bucketVersioningV2State struct {
	// Name of the S3 bucket.
	Bucket *string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.
	Mfa *string `pulumi:"mfa"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Configuration block for the versioning parameters. See below.
	VersioningConfiguration *BucketVersioningV2VersioningConfiguration `pulumi:"versioningConfiguration"`
}

type BucketVersioningV2State struct {
	// Name of the S3 bucket.
	Bucket pulumi.StringPtrInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.
	Mfa pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Configuration block for the versioning parameters. See below.
	VersioningConfiguration BucketVersioningV2VersioningConfigurationPtrInput
}

func (BucketVersioningV2State) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketVersioningV2State)(nil)).Elem()
}

type bucketVersioningV2Args struct {
	// Name of the S3 bucket.
	Bucket string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.
	Mfa *string `pulumi:"mfa"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Configuration block for the versioning parameters. See below.
	VersioningConfiguration BucketVersioningV2VersioningConfiguration `pulumi:"versioningConfiguration"`
}

// The set of arguments for constructing a BucketVersioningV2 resource.
type BucketVersioningV2Args struct {
	// Name of the S3 bucket.
	Bucket pulumi.StringInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.
	Mfa pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Configuration block for the versioning parameters. See below.
	VersioningConfiguration BucketVersioningV2VersioningConfigurationInput
}

func (BucketVersioningV2Args) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketVersioningV2Args)(nil)).Elem()
}

type BucketVersioningV2Input interface {
	pulumi.Input

	ToBucketVersioningV2Output() BucketVersioningV2Output
	ToBucketVersioningV2OutputWithContext(ctx context.Context) BucketVersioningV2Output
}

func (*BucketVersioningV2) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketVersioningV2)(nil)).Elem()
}

func (i *BucketVersioningV2) ToBucketVersioningV2Output() BucketVersioningV2Output {
	return i.ToBucketVersioningV2OutputWithContext(context.Background())
}

func (i *BucketVersioningV2) ToBucketVersioningV2OutputWithContext(ctx context.Context) BucketVersioningV2Output {
	return pulumi.ToOutputWithContext(ctx, i).(BucketVersioningV2Output)
}

// BucketVersioningV2ArrayInput is an input type that accepts BucketVersioningV2Array and BucketVersioningV2ArrayOutput values.
// You can construct a concrete instance of `BucketVersioningV2ArrayInput` via:
//
//	BucketVersioningV2Array{ BucketVersioningV2Args{...} }
type BucketVersioningV2ArrayInput interface {
	pulumi.Input

	ToBucketVersioningV2ArrayOutput() BucketVersioningV2ArrayOutput
	ToBucketVersioningV2ArrayOutputWithContext(context.Context) BucketVersioningV2ArrayOutput
}

type BucketVersioningV2Array []BucketVersioningV2Input

func (BucketVersioningV2Array) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketVersioningV2)(nil)).Elem()
}

func (i BucketVersioningV2Array) ToBucketVersioningV2ArrayOutput() BucketVersioningV2ArrayOutput {
	return i.ToBucketVersioningV2ArrayOutputWithContext(context.Background())
}

func (i BucketVersioningV2Array) ToBucketVersioningV2ArrayOutputWithContext(ctx context.Context) BucketVersioningV2ArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketVersioningV2ArrayOutput)
}

// BucketVersioningV2MapInput is an input type that accepts BucketVersioningV2Map and BucketVersioningV2MapOutput values.
// You can construct a concrete instance of `BucketVersioningV2MapInput` via:
//
//	BucketVersioningV2Map{ "key": BucketVersioningV2Args{...} }
type BucketVersioningV2MapInput interface {
	pulumi.Input

	ToBucketVersioningV2MapOutput() BucketVersioningV2MapOutput
	ToBucketVersioningV2MapOutputWithContext(context.Context) BucketVersioningV2MapOutput
}

type BucketVersioningV2Map map[string]BucketVersioningV2Input

func (BucketVersioningV2Map) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketVersioningV2)(nil)).Elem()
}

func (i BucketVersioningV2Map) ToBucketVersioningV2MapOutput() BucketVersioningV2MapOutput {
	return i.ToBucketVersioningV2MapOutputWithContext(context.Background())
}

func (i BucketVersioningV2Map) ToBucketVersioningV2MapOutputWithContext(ctx context.Context) BucketVersioningV2MapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketVersioningV2MapOutput)
}

type BucketVersioningV2Output struct{ *pulumi.OutputState }

func (BucketVersioningV2Output) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketVersioningV2)(nil)).Elem()
}

func (o BucketVersioningV2Output) ToBucketVersioningV2Output() BucketVersioningV2Output {
	return o
}

func (o BucketVersioningV2Output) ToBucketVersioningV2OutputWithContext(ctx context.Context) BucketVersioningV2Output {
	return o
}

// Name of the S3 bucket.
func (o BucketVersioningV2Output) Bucket() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketVersioningV2) pulumi.StringOutput { return v.Bucket }).(pulumi.StringOutput)
}

// Account ID of the expected bucket owner.
func (o BucketVersioningV2Output) ExpectedBucketOwner() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *BucketVersioningV2) pulumi.StringPtrOutput { return v.ExpectedBucketOwner }).(pulumi.StringPtrOutput)
}

// Concatenation of the authentication device's serial number, a space, and the value that is displayed on your authentication device.
func (o BucketVersioningV2Output) Mfa() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *BucketVersioningV2) pulumi.StringPtrOutput { return v.Mfa }).(pulumi.StringPtrOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o BucketVersioningV2Output) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketVersioningV2) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// Configuration block for the versioning parameters. See below.
func (o BucketVersioningV2Output) VersioningConfiguration() BucketVersioningV2VersioningConfigurationOutput {
	return o.ApplyT(func(v *BucketVersioningV2) BucketVersioningV2VersioningConfigurationOutput {
		return v.VersioningConfiguration
	}).(BucketVersioningV2VersioningConfigurationOutput)
}

type BucketVersioningV2ArrayOutput struct{ *pulumi.OutputState }

func (BucketVersioningV2ArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketVersioningV2)(nil)).Elem()
}

func (o BucketVersioningV2ArrayOutput) ToBucketVersioningV2ArrayOutput() BucketVersioningV2ArrayOutput {
	return o
}

func (o BucketVersioningV2ArrayOutput) ToBucketVersioningV2ArrayOutputWithContext(ctx context.Context) BucketVersioningV2ArrayOutput {
	return o
}

func (o BucketVersioningV2ArrayOutput) Index(i pulumi.IntInput) BucketVersioningV2Output {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *BucketVersioningV2 {
		return vs[0].([]*BucketVersioningV2)[vs[1].(int)]
	}).(BucketVersioningV2Output)
}

type BucketVersioningV2MapOutput struct{ *pulumi.OutputState }

func (BucketVersioningV2MapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketVersioningV2)(nil)).Elem()
}

func (o BucketVersioningV2MapOutput) ToBucketVersioningV2MapOutput() BucketVersioningV2MapOutput {
	return o
}

func (o BucketVersioningV2MapOutput) ToBucketVersioningV2MapOutputWithContext(ctx context.Context) BucketVersioningV2MapOutput {
	return o
}

func (o BucketVersioningV2MapOutput) MapIndex(k pulumi.StringInput) BucketVersioningV2Output {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *BucketVersioningV2 {
		return vs[0].(map[string]*BucketVersioningV2)[vs[1].(string)]
	}).(BucketVersioningV2Output)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*BucketVersioningV2Input)(nil)).Elem(), &BucketVersioningV2{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketVersioningV2ArrayInput)(nil)).Elem(), BucketVersioningV2Array{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketVersioningV2MapInput)(nil)).Elem(), BucketVersioningV2Map{})
	pulumi.RegisterOutputType(BucketVersioningV2Output{})
	pulumi.RegisterOutputType(BucketVersioningV2ArrayOutput{})
	pulumi.RegisterOutputType(BucketVersioningV2MapOutput{})
}
