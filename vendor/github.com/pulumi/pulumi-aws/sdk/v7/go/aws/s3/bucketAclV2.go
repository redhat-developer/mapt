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

// Provides an S3 bucket ACL resource.
//
// > **Note:** destroy does not delete the S3 Bucket ACL but does remove the resource from state.
//
// > This resource cannot be used with S3 directory buckets.
//
// ## Example Usage
//
// ### With `private` ACL
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
//				Bucket: pulumi.String("my-tf-example-bucket"),
//			})
//			if err != nil {
//				return err
//			}
//			exampleBucketOwnershipControls, err := s3.NewBucketOwnershipControls(ctx, "example", &s3.BucketOwnershipControlsArgs{
//				Bucket: example.ID(),
//				Rule: &s3.BucketOwnershipControlsRuleArgs{
//					ObjectOwnership: pulumi.String("BucketOwnerPreferred"),
//				},
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketAcl(ctx, "example", &s3.BucketAclArgs{
//				Bucket: example.ID(),
//				Acl:    pulumi.String("private"),
//			}, pulumi.DependsOn([]pulumi.Resource{
//				exampleBucketOwnershipControls,
//			}))
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
//
// ### With `public-read` ACL
//
// > This example explicitly disables the default S3 bucket security settings. This
// should be done with caution, as all bucket objects become publicly exposed.
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
//				Bucket: pulumi.String("my-tf-example-bucket"),
//			})
//			if err != nil {
//				return err
//			}
//			exampleBucketOwnershipControls, err := s3.NewBucketOwnershipControls(ctx, "example", &s3.BucketOwnershipControlsArgs{
//				Bucket: example.ID(),
//				Rule: &s3.BucketOwnershipControlsRuleArgs{
//					ObjectOwnership: pulumi.String("BucketOwnerPreferred"),
//				},
//			})
//			if err != nil {
//				return err
//			}
//			exampleBucketPublicAccessBlock, err := s3.NewBucketPublicAccessBlock(ctx, "example", &s3.BucketPublicAccessBlockArgs{
//				Bucket:                example.ID(),
//				BlockPublicAcls:       pulumi.Bool(false),
//				BlockPublicPolicy:     pulumi.Bool(false),
//				IgnorePublicAcls:      pulumi.Bool(false),
//				RestrictPublicBuckets: pulumi.Bool(false),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketAcl(ctx, "example", &s3.BucketAclArgs{
//				Bucket: example.ID(),
//				Acl:    pulumi.String("public-read"),
//			}, pulumi.DependsOn([]pulumi.Resource{
//				exampleBucketOwnershipControls,
//				exampleBucketPublicAccessBlock,
//			}))
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
//
// ### With Grants
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
//			current, err := s3.GetCanonicalUserId(ctx, map[string]interface{}{}, nil)
//			if err != nil {
//				return err
//			}
//			example, err := s3.NewBucket(ctx, "example", &s3.BucketArgs{
//				Bucket: pulumi.String("my-tf-example-bucket"),
//			})
//			if err != nil {
//				return err
//			}
//			exampleBucketOwnershipControls, err := s3.NewBucketOwnershipControls(ctx, "example", &s3.BucketOwnershipControlsArgs{
//				Bucket: example.ID(),
//				Rule: &s3.BucketOwnershipControlsRuleArgs{
//					ObjectOwnership: pulumi.String("BucketOwnerPreferred"),
//				},
//			})
//			if err != nil {
//				return err
//			}
//			_, err = s3.NewBucketAcl(ctx, "example", &s3.BucketAclArgs{
//				Bucket: example.ID(),
//				AccessControlPolicy: &s3.BucketAclAccessControlPolicyArgs{
//					Grants: s3.BucketAclAccessControlPolicyGrantArray{
//						&s3.BucketAclAccessControlPolicyGrantArgs{
//							Grantee: &s3.BucketAclAccessControlPolicyGrantGranteeArgs{
//								Id:   pulumi.String(current.Id),
//								Type: pulumi.String("CanonicalUser"),
//							},
//							Permission: pulumi.String("READ"),
//						},
//						&s3.BucketAclAccessControlPolicyGrantArgs{
//							Grantee: &s3.BucketAclAccessControlPolicyGrantGranteeArgs{
//								Type: pulumi.String("Group"),
//								Uri:  pulumi.String("http://acs.amazonaws.com/groups/s3/LogDelivery"),
//							},
//							Permission: pulumi.String("READ_ACP"),
//						},
//					},
//					Owner: &s3.BucketAclAccessControlPolicyOwnerArgs{
//						Id: pulumi.String(current.Id),
//					},
//				},
//			}, pulumi.DependsOn([]pulumi.Resource{
//				exampleBucketOwnershipControls,
//			}))
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
// If the owner (account ID) of the source bucket is the _same_ account used to configure the AWS Provider, and the source bucket is __configured__ with a
// [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), import using the `bucket` and `acl` separated by a comma (`,`):
//
// If the owner (account ID) of the source bucket _differs_ from the account used to configure the AWS Provider, and the source bucket is __not configured__ with a [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// If the owner (account ID) of the source bucket _differs_ from the account used to configure the AWS Provider, and the source bucket is __configured__ with a
// [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), imported using the `bucket`, `expected_bucket_owner`, and `acl` separated by commas (`,`):
//
// __Using `pulumi import` to import__ using `bucket`, `expected_bucket_owner`, and/or `acl`, depending on your situation. For example:
//
// If the owner (account ID) of the source bucket is the _same_ account used to configure the AWS Provider, and the source bucket is __not configured__ with a
// [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), import using the `bucket`:
//
// ```sh
// $ pulumi import aws:s3/bucketAclV2:BucketAclV2 example bucket-name
// ```
// If the owner (account ID) of the source bucket is the _same_ account used to configure the AWS Provider, and the source bucket is __configured__ with a [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), import using the `bucket` and `acl` separated by a comma (`,`):
//
// ```sh
// $ pulumi import aws:s3/bucketAclV2:BucketAclV2 example bucket-name,private
// ```
// If the owner (account ID) of the source bucket _differs_ from the account used to configure the AWS Provider, and the source bucket is __not configured__ with a [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):
//
// ```sh
// $ pulumi import aws:s3/bucketAclV2:BucketAclV2 example bucket-name,123456789012
// ```
// If the owner (account ID) of the source bucket _differs_ from the account used to configure the AWS Provider, and the source bucket is __configured__ with a [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) (i.e. predefined grant), imported using the `bucket`, `expected_bucket_owner`, and `acl` separated by commas (`,`):
//
// ```sh
// $ pulumi import aws:s3/bucketAclV2:BucketAclV2 example bucket-name,123456789012,private
// ```
//
// Deprecated: aws.s3/bucketaclv2.BucketAclV2 has been deprecated in favor of aws.s3/bucketacl.BucketAcl
type BucketAclV2 struct {
	pulumi.CustomResourceState

	// Configuration block that sets the ACL permissions for an object per grantee. See below.
	AccessControlPolicy BucketAclV2AccessControlPolicyOutput `pulumi:"accessControlPolicy"`
	// Specifies the Canned ACL to apply to the bucket. Valid values: `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `bucket-owner-full-control`, `log-delivery-write`. Full details are available on the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl).
	Acl pulumi.StringPtrOutput `pulumi:"acl"`
	// Bucket to which to apply the ACL.
	Bucket pulumi.StringOutput `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrOutput `pulumi:"expectedBucketOwner"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
}

// NewBucketAclV2 registers a new resource with the given unique name, arguments, and options.
func NewBucketAclV2(ctx *pulumi.Context,
	name string, args *BucketAclV2Args, opts ...pulumi.ResourceOption) (*BucketAclV2, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Bucket == nil {
		return nil, errors.New("invalid value for required argument 'Bucket'")
	}
	aliases := pulumi.Aliases([]pulumi.Alias{
		{
			Type: pulumi.String("aws:s3/bucketAclV2:BucketAclV2"),
		},
	})
	opts = append(opts, aliases)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource BucketAclV2
	err := ctx.RegisterResource("aws:s3/bucketAclV2:BucketAclV2", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetBucketAclV2 gets an existing BucketAclV2 resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetBucketAclV2(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *BucketAclV2State, opts ...pulumi.ResourceOption) (*BucketAclV2, error) {
	var resource BucketAclV2
	err := ctx.ReadResource("aws:s3/bucketAclV2:BucketAclV2", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering BucketAclV2 resources.
type bucketAclV2State struct {
	// Configuration block that sets the ACL permissions for an object per grantee. See below.
	AccessControlPolicy *BucketAclV2AccessControlPolicy `pulumi:"accessControlPolicy"`
	// Specifies the Canned ACL to apply to the bucket. Valid values: `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `bucket-owner-full-control`, `log-delivery-write`. Full details are available on the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl).
	Acl *string `pulumi:"acl"`
	// Bucket to which to apply the ACL.
	Bucket *string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

type BucketAclV2State struct {
	// Configuration block that sets the ACL permissions for an object per grantee. See below.
	AccessControlPolicy BucketAclV2AccessControlPolicyPtrInput
	// Specifies the Canned ACL to apply to the bucket. Valid values: `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `bucket-owner-full-control`, `log-delivery-write`. Full details are available on the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl).
	Acl pulumi.StringPtrInput
	// Bucket to which to apply the ACL.
	Bucket pulumi.StringPtrInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
}

func (BucketAclV2State) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketAclV2State)(nil)).Elem()
}

type bucketAclV2Args struct {
	// Configuration block that sets the ACL permissions for an object per grantee. See below.
	AccessControlPolicy *BucketAclV2AccessControlPolicy `pulumi:"accessControlPolicy"`
	// Specifies the Canned ACL to apply to the bucket. Valid values: `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `bucket-owner-full-control`, `log-delivery-write`. Full details are available on the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl).
	Acl *string `pulumi:"acl"`
	// Bucket to which to apply the ACL.
	Bucket string `pulumi:"bucket"`
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner *string `pulumi:"expectedBucketOwner"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

// The set of arguments for constructing a BucketAclV2 resource.
type BucketAclV2Args struct {
	// Configuration block that sets the ACL permissions for an object per grantee. See below.
	AccessControlPolicy BucketAclV2AccessControlPolicyPtrInput
	// Specifies the Canned ACL to apply to the bucket. Valid values: `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `bucket-owner-full-control`, `log-delivery-write`. Full details are available on the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl).
	Acl pulumi.StringPtrInput
	// Bucket to which to apply the ACL.
	Bucket pulumi.StringInput
	// Account ID of the expected bucket owner.
	ExpectedBucketOwner pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
}

func (BucketAclV2Args) ElementType() reflect.Type {
	return reflect.TypeOf((*bucketAclV2Args)(nil)).Elem()
}

type BucketAclV2Input interface {
	pulumi.Input

	ToBucketAclV2Output() BucketAclV2Output
	ToBucketAclV2OutputWithContext(ctx context.Context) BucketAclV2Output
}

func (*BucketAclV2) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketAclV2)(nil)).Elem()
}

func (i *BucketAclV2) ToBucketAclV2Output() BucketAclV2Output {
	return i.ToBucketAclV2OutputWithContext(context.Background())
}

func (i *BucketAclV2) ToBucketAclV2OutputWithContext(ctx context.Context) BucketAclV2Output {
	return pulumi.ToOutputWithContext(ctx, i).(BucketAclV2Output)
}

// BucketAclV2ArrayInput is an input type that accepts BucketAclV2Array and BucketAclV2ArrayOutput values.
// You can construct a concrete instance of `BucketAclV2ArrayInput` via:
//
//	BucketAclV2Array{ BucketAclV2Args{...} }
type BucketAclV2ArrayInput interface {
	pulumi.Input

	ToBucketAclV2ArrayOutput() BucketAclV2ArrayOutput
	ToBucketAclV2ArrayOutputWithContext(context.Context) BucketAclV2ArrayOutput
}

type BucketAclV2Array []BucketAclV2Input

func (BucketAclV2Array) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketAclV2)(nil)).Elem()
}

func (i BucketAclV2Array) ToBucketAclV2ArrayOutput() BucketAclV2ArrayOutput {
	return i.ToBucketAclV2ArrayOutputWithContext(context.Background())
}

func (i BucketAclV2Array) ToBucketAclV2ArrayOutputWithContext(ctx context.Context) BucketAclV2ArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketAclV2ArrayOutput)
}

// BucketAclV2MapInput is an input type that accepts BucketAclV2Map and BucketAclV2MapOutput values.
// You can construct a concrete instance of `BucketAclV2MapInput` via:
//
//	BucketAclV2Map{ "key": BucketAclV2Args{...} }
type BucketAclV2MapInput interface {
	pulumi.Input

	ToBucketAclV2MapOutput() BucketAclV2MapOutput
	ToBucketAclV2MapOutputWithContext(context.Context) BucketAclV2MapOutput
}

type BucketAclV2Map map[string]BucketAclV2Input

func (BucketAclV2Map) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketAclV2)(nil)).Elem()
}

func (i BucketAclV2Map) ToBucketAclV2MapOutput() BucketAclV2MapOutput {
	return i.ToBucketAclV2MapOutputWithContext(context.Background())
}

func (i BucketAclV2Map) ToBucketAclV2MapOutputWithContext(ctx context.Context) BucketAclV2MapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(BucketAclV2MapOutput)
}

type BucketAclV2Output struct{ *pulumi.OutputState }

func (BucketAclV2Output) ElementType() reflect.Type {
	return reflect.TypeOf((**BucketAclV2)(nil)).Elem()
}

func (o BucketAclV2Output) ToBucketAclV2Output() BucketAclV2Output {
	return o
}

func (o BucketAclV2Output) ToBucketAclV2OutputWithContext(ctx context.Context) BucketAclV2Output {
	return o
}

// Configuration block that sets the ACL permissions for an object per grantee. See below.
func (o BucketAclV2Output) AccessControlPolicy() BucketAclV2AccessControlPolicyOutput {
	return o.ApplyT(func(v *BucketAclV2) BucketAclV2AccessControlPolicyOutput { return v.AccessControlPolicy }).(BucketAclV2AccessControlPolicyOutput)
}

// Specifies the Canned ACL to apply to the bucket. Valid values: `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, `bucket-owner-read`, `bucket-owner-full-control`, `log-delivery-write`. Full details are available on the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl).
func (o BucketAclV2Output) Acl() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *BucketAclV2) pulumi.StringPtrOutput { return v.Acl }).(pulumi.StringPtrOutput)
}

// Bucket to which to apply the ACL.
func (o BucketAclV2Output) Bucket() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketAclV2) pulumi.StringOutput { return v.Bucket }).(pulumi.StringOutput)
}

// Account ID of the expected bucket owner.
func (o BucketAclV2Output) ExpectedBucketOwner() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *BucketAclV2) pulumi.StringPtrOutput { return v.ExpectedBucketOwner }).(pulumi.StringPtrOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o BucketAclV2Output) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *BucketAclV2) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

type BucketAclV2ArrayOutput struct{ *pulumi.OutputState }

func (BucketAclV2ArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*BucketAclV2)(nil)).Elem()
}

func (o BucketAclV2ArrayOutput) ToBucketAclV2ArrayOutput() BucketAclV2ArrayOutput {
	return o
}

func (o BucketAclV2ArrayOutput) ToBucketAclV2ArrayOutputWithContext(ctx context.Context) BucketAclV2ArrayOutput {
	return o
}

func (o BucketAclV2ArrayOutput) Index(i pulumi.IntInput) BucketAclV2Output {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *BucketAclV2 {
		return vs[0].([]*BucketAclV2)[vs[1].(int)]
	}).(BucketAclV2Output)
}

type BucketAclV2MapOutput struct{ *pulumi.OutputState }

func (BucketAclV2MapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*BucketAclV2)(nil)).Elem()
}

func (o BucketAclV2MapOutput) ToBucketAclV2MapOutput() BucketAclV2MapOutput {
	return o
}

func (o BucketAclV2MapOutput) ToBucketAclV2MapOutputWithContext(ctx context.Context) BucketAclV2MapOutput {
	return o
}

func (o BucketAclV2MapOutput) MapIndex(k pulumi.StringInput) BucketAclV2Output {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *BucketAclV2 {
		return vs[0].(map[string]*BucketAclV2)[vs[1].(string)]
	}).(BucketAclV2Output)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*BucketAclV2Input)(nil)).Elem(), &BucketAclV2{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketAclV2ArrayInput)(nil)).Elem(), BucketAclV2Array{})
	pulumi.RegisterInputType(reflect.TypeOf((*BucketAclV2MapInput)(nil)).Elem(), BucketAclV2Map{})
	pulumi.RegisterOutputType(BucketAclV2Output{})
	pulumi.RegisterOutputType(BucketAclV2ArrayOutput{})
	pulumi.RegisterOutputType(BucketAclV2MapOutput{})
}
