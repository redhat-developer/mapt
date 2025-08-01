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

// Provides a S3 bucket server-side encryption configuration resource.
//
// > **NOTE:** Destroying an `s3.BucketServerSideEncryptionConfiguration` resource resets the bucket to [Amazon S3 bucket default encryption](https://docs.aws.amazon.com/AmazonS3/latest/userguide/default-encryption-faq.html).
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kms"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			mykey, err := kms.NewKey(ctx, "mykey", &kms.KeyArgs{
//				Description:          pulumi.String("This key is used to encrypt bucket objects"),
//				DeletionWindowInDays: pulumi.Int(10),
//			})
//			if err != nil {
//				return err
//			}
//			mybucket, err := s3.NewBucket(ctx, "mybucket", &s3.BucketArgs{
//				Bucket: pulumi.String("mybucket"),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketServerSideEncryptionConfiguration(ctx, "example", &s3.BucketServerSideEncryptionConfigurationArgs{
//				Bucket: mybucket.ID(),
//				Rules: s3.BucketServerSideEncryptionConfigurationRuleArray{
//					&s3.BucketServerSideEncryptionConfigurationRuleArgs{
//						ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
//							KmsMasterKeyId: mykey.Arn,
//							SseAlgorithm:   pulumi.String("aws:kms"),
//						},
//					},
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
// ## Import
//
// If the owner (account ID) of the source bucket differs from the account used to configure the AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// __Using `pulumi import` to import__ S3 bucket server-side encryption configuration using the `bucket` or using the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:
//
// If the owner (account ID) of the source bucket is the same account used to configure the AWS Provider, import using the `bucket`:
//
// ```sh
// $ pulumi import aws:s3/bucketServerSideEncryptionConfiguration:BucketServerSideEncryptionConfiguration example bucket-name
// ```
// If the owner (account ID) of the source bucket differs from the account used to configure the AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// ```sh
// $ pulumi import aws:s3/bucketServerSideEncryptionConfiguration:BucketServerSideEncryptionConfiguration example bucket-name,123456789012
// ```
type BucketServerSideEncryptionConfiguration struct {
	pulumi.CustomResourceState

	// ID (name) of the bucket.
	Bucket pulumi.StringOutput `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrOutput `pulumi:"expectedBucketOwner"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// Set of server-side encryption configuration rules. See below. Currently, only a single rule is supported.
	Rules BucketServerSideEncryptionConfigurationRuleArrayOutput `pulumi:"rules"`
}

// NewBucketServerSideEncryptionConfiguration registers a new resource with the given unique name, arguments, and options.
func NewBucketServerSideEncryptionConfiguration(ctx *pulumi.Context,
	name string, args *BucketServerSideEncryptionConfigurationArgs, opts ...pulumi.ResourceOption) (*BucketServerSideEncryptionConfiguration, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Bucket == nil {
		return nil, errors.New("invalid value for required argument 'Bucket'")
	}
	if args.Rules == nil {
		return nil, errors.New("invalid value for required argument 'Rules'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("aws:s3/bucketServerSideEncryptionConfigurationV2:BucketServerSideEncryptionConfigurationV2"),
		},
	})
	opts = append(opts, aliases)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource BucketServerSideEncryptionConfiguration
	err := ctx.RegisterResource("aws:s3/bucketServerSideEncryptionConfiguration:BucketServerSideEncryptionConfiguration", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetBucketServerSideEncryptionConfiguration gets an existing BucketServerSideEncryptionConfiguration resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetBucketServerSideEncryptionConfiguration(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *BucketServerSideEncryptionConfigurationState, opts ...pulumi.ResourceOption) (*BucketServerSideEncryptionConfiguration, error) {
	var resource BucketServerSideEncryptionConfiguration
	err := ctx.ReadResource("aws:s3/bucketServerSideEncryptionConfiguration:BucketServerSideEncryptionConfiguration", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering BucketServerSideEncryptionConfiguration resources.
type bucketServerSideEncryptionConfigurationState struct {
	// ID (name) of the bucket.
	Bucket *string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Set of server-side encryption configuration rules. See below. Currently, only a single rule is supported.
	Rules []BucketServerSideEncryptionConfigurationRule `pulumi:"rules"`
}

type BucketServerSideEncryptionConfigurationState struct {
	// ID (name) of the bucket.
	Bucket pulumi.StringPtrInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Set of server-side encryption configuration rules. See below. Currently, only a single rule is supported.
	Rules BucketServerSideEncryptionConfigurationRuleArrayInput
}

func (BucketServerSideEncryptionConfigurationState) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketServerSideEncryptionConfigurationState)(nil)).Elem()
}

type bucketServerSideEncryptionConfigurationArgs struct {
	// ID (name) of the bucket.
	Bucket string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Set of server-side encryption configuration rules. See below. Currently, only a single rule is supported.
	Rules []BucketServerSideEncryptionConfigurationRule `pulumi:"rules"`
}

// The set of arguments for constructing a BucketServerSideEncryptionConfiguration resource.
type BucketServerSideEncryptionConfigurationArgs struct {
	// ID (name) of the bucket.
	Bucket pulumi.StringInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Set of server-side encryption configuration rules. See below. Currently, only a single rule is supported.
	Rules BucketServerSideEncryptionConfigurationRuleArrayInput
}

func (BucketServerSideEncryptionConfigurationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketServerSideEncryptionConfigurationArgs)(nil)).Elem()
}

type BucketServerSideEncryptionConfigurationInput interface {
	pulumi.Input

	ToBucketServerSideEncryptionConfigurationOutput() BucketServerSideEncryptionConfigurationOutput
	ToBucketServerSideEncryptionConfigurationOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationOutput
}

func (*BucketServerSideEncryptionConfiguration) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketServerSideEncryptionConfiguration)(nil)).Elem()
}

func (i *BucketServerSideEncryptionConfiguration) ToBucketServerSideEncryptionConfigurationOutput() BucketServerSideEncryptionConfigurationOutput {
	return i.ToBucketServerSideEncryptionConfigurationOutputWithContext(context.Background())
}

func (i *BucketServerSideEncryptionConfiguration) ToBucketServerSideEncryptionConfigurationOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketServerSideEncryptionConfigurationOutput)
}

// BucketServerSideEncryptionConfigurationArrayInput is an input type that accepts BucketServerSideEncryptionConfigurationArray and BucketServerSideEncryptionConfigurationArrayOutput values.
// You can construct a concrete instance of `BucketServerSideEncryptionConfigurationArrayInput` via:
//
//	BucketServerSideEncryptionConfigurationArray{ BucketServerSideEncryptionConfigurationArgs{...} }
type BucketServerSideEncryptionConfigurationArrayInput interface {
	pulumi.Input

	ToBucketServerSideEncryptionConfigurationArrayOutput() BucketServerSideEncryptionConfigurationArrayOutput
	ToBucketServerSideEncryptionConfigurationArrayOutputWithContext(context.Context) BucketServerSideEncryptionConfigurationArrayOutput
}

type BucketServerSideEncryptionConfigurationArray []BucketServerSideEncryptionConfigurationInput

func (BucketServerSideEncryptionConfigurationArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketServerSideEncryptionConfiguration)(nil)).Elem()
}

func (i BucketServerSideEncryptionConfigurationArray) ToBucketServerSideEncryptionConfigurationArrayOutput() BucketServerSideEncryptionConfigurationArrayOutput {
	return i.ToBucketServerSideEncryptionConfigurationArrayOutputWithContext(context.Background())
}

func (i BucketServerSideEncryptionConfigurationArray) ToBucketServerSideEncryptionConfigurationArrayOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketServerSideEncryptionConfigurationArrayOutput)
}

// BucketServerSideEncryptionConfigurationMapInput is an input type that accepts BucketServerSideEncryptionConfigurationMap and BucketServerSideEncryptionConfigurationMapOutput values.
// You can construct a concrete instance of `BucketServerSideEncryptionConfigurationMapInput` via:
//
//	BucketServerSideEncryptionConfigurationMap{ "key": BucketServerSideEncryptionConfigurationArgs{...} }
type BucketServerSideEncryptionConfigurationMapInput interface {
	pulumi.Input

	ToBucketServerSideEncryptionConfigurationMapOutput() BucketServerSideEncryptionConfigurationMapOutput
	ToBucketServerSideEncryptionConfigurationMapOutputWithContext(context.Context) BucketServerSideEncryptionConfigurationMapOutput
}

type BucketServerSideEncryptionConfigurationMap map[string]BucketServerSideEncryptionConfigurationInput

func (BucketServerSideEncryptionConfigurationMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketServerSideEncryptionConfiguration)(nil)).Elem()
}

func (i BucketServerSideEncryptionConfigurationMap) ToBucketServerSideEncryptionConfigurationMapOutput() BucketServerSideEncryptionConfigurationMapOutput {
	return i.ToBucketServerSideEncryptionConfigurationMapOutputWithContext(context.Background())
}

func (i BucketServerSideEncryptionConfigurationMap) ToBucketServerSideEncryptionConfigurationMapOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketServerSideEncryptionConfigurationMapOutput)
}

type BucketServerSideEncryptionConfigurationOutput struct{ *pulumi.OutputState }

func (BucketServerSideEncryptionConfigurationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketServerSideEncryptionConfiguration)(nil)).Elem()
}

func (o BucketServerSideEncryptionConfigurationOutput) ToBucketServerSideEncryptionConfigurationOutput() BucketServerSideEncryptionConfigurationOutput {
	return o
}

func (o BucketServerSideEncryptionConfigurationOutput) ToBucketServerSideEncryptionConfigurationOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationOutput {
	return o
}

// ID (name) of the bucket.
func (o BucketServerSideEncryptionConfigurationOutput) Bucket() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketServerSideEncryptionConfiguration) pulumi.StringOutput { return v.Bucket }).(pulumi.StringOutput)
}

// Account ID of the expected bucket owner.
func (o BucketServerSideEncryptionConfigurationOutput) ExpectedBucketOwner() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *BucketServerSideEncryptionConfiguration) pulumi.StringPtrOutput { return v.ExpectedBucketOwner }).(pulumi.StringPtrOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o BucketServerSideEncryptionConfigurationOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketServerSideEncryptionConfiguration) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// Set of server-side encryption configuration rules. See below. Currently, only a single rule is supported.
func (o BucketServerSideEncryptionConfigurationOutput) Rules() BucketServerSideEncryptionConfigurationRuleArrayOutput {
	return o.ApplyT(func(v *BucketServerSideEncryptionConfiguration) BucketServerSideEncryptionConfigurationRuleArrayOutput {
		return v.Rules
	}).(BucketServerSideEncryptionConfigurationRuleArrayOutput)
}

type BucketServerSideEncryptionConfigurationArrayOutput struct{ *pulumi.OutputState }

func (BucketServerSideEncryptionConfigurationArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketServerSideEncryptionConfiguration)(nil)).Elem()
}

func (o BucketServerSideEncryptionConfigurationArrayOutput) ToBucketServerSideEncryptionConfigurationArrayOutput() BucketServerSideEncryptionConfigurationArrayOutput {
	return o
}

func (o BucketServerSideEncryptionConfigurationArrayOutput) ToBucketServerSideEncryptionConfigurationArrayOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationArrayOutput {
	return o
}

func (o BucketServerSideEncryptionConfigurationArrayOutput) Index(i pulumi.IntInput) BucketServerSideEncryptionConfigurationOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *BucketServerSideEncryptionConfiguration {
		return vs[0].([]*BucketServerSideEncryptionConfiguration)[vs[1].(int)]
	}).(BucketServerSideEncryptionConfigurationOutput)
}

type BucketServerSideEncryptionConfigurationMapOutput struct{ *pulumi.OutputState }

func (BucketServerSideEncryptionConfigurationMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketServerSideEncryptionConfiguration)(nil)).Elem()
}

func (o BucketServerSideEncryptionConfigurationMapOutput) ToBucketServerSideEncryptionConfigurationMapOutput() BucketServerSideEncryptionConfigurationMapOutput {
	return o
}

func (o BucketServerSideEncryptionConfigurationMapOutput) ToBucketServerSideEncryptionConfigurationMapOutputWithContext(ctx context.Context) BucketServerSideEncryptionConfigurationMapOutput {
	return o
}

func (o BucketServerSideEncryptionConfigurationMapOutput) MapIndex(k pulumi.StringInput) BucketServerSideEncryptionConfigurationOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *BucketServerSideEncryptionConfiguration {
		return vs[0].(map[string]*BucketServerSideEncryptionConfiguration)[vs[1].(string)]
	}).(BucketServerSideEncryptionConfigurationOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*BucketServerSideEncryptionConfigurationInput)(nil)).Elem(), &BucketServerSideEncryptionConfiguration{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketServerSideEncryptionConfigurationArrayInput)(nil)).Elem(), BucketServerSideEncryptionConfigurationArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketServerSideEncryptionConfigurationMapInput)(nil)).Elem(), BucketServerSideEncryptionConfigurationMap{})
	pulumi.RegisterOutputType(BucketServerSideEncryptionConfigurationOutput{})
	pulumi.RegisterOutputType(BucketServerSideEncryptionConfigurationArrayOutput{})
	pulumi.RegisterOutputType(BucketServerSideEncryptionConfigurationMapOutput{})
}
