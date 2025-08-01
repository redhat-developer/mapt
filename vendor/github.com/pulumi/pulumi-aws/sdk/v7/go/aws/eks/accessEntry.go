// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package eks

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Access Entry Configurations for an EKS Cluster.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := eks.NewAccessEntry(ctx, "example", &eks.AccessEntryArgs{
//				ClusterName:  pulumi.Any(exampleAwsEksCluster.Name),
//				PrincipalArn: pulumi.Any(exampleAwsIamRole.Arn),
//				KubernetesGroups: pulumi.StringArray{
//					pulumi.String("group-1"),
//					pulumi.String("group-2"),
//				},
//				Type: pulumi.String("STANDARD"),
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
// Using `pulumi import`, import EKS access entry using the `cluster_name` and `principal_arn` separated by a colon (`:`). For example:
//
// ```sh
// $ pulumi import aws:eks/accessEntry:AccessEntry my_eks_access_entry my_cluster_name:my_principal_arn
// ```
type AccessEntry struct {
	pulumi.CustomResourceState

	// Amazon Resource Name (ARN) of the Access Entry.
	AccessEntryArn pulumi.StringOutput `pulumi:"accessEntryArn"`
	// Name of the EKS Cluster.
	ClusterName pulumi.StringOutput `pulumi:"clusterName"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
	CreatedAt pulumi.StringOutput `pulumi:"createdAt"`
	// List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
	KubernetesGroups pulumi.StringArrayOutput `pulumi:"kubernetesGroups"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
	ModifiedAt pulumi.StringOutput `pulumi:"modifiedAt"`
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	//
	// The following arguments are optional:
	PrincipalArn pulumi.StringOutput `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// (Optional) Key-value map of resource tags, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapOutput `pulumi:"tagsAll"`
	// Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
	Type pulumi.StringPtrOutput `pulumi:"type"`
	// Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
	UserName pulumi.StringOutput `pulumi:"userName"`
}

// NewAccessEntry registers a new resource with the given unique name, arguments, and options.
func NewAccessEntry(ctx *pulumi.Context,
	name string, args *AccessEntryArgs, opts ...pulumi.ResourceOption) (*AccessEntry, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.ClusterName == nil {
		return nil, errors.New("invalid value for required argument 'ClusterName'")
	}
	if args.PrincipalArn == nil {
		return nil, errors.New("invalid value for required argument 'PrincipalArn'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource AccessEntry
	err := ctx.RegisterResource("aws:eks/accessEntry:AccessEntry", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetAccessEntry gets an existing AccessEntry resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetAccessEntry(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *AccessEntryState, opts ...pulumi.ResourceOption) (*AccessEntry, error) {
	var resource AccessEntry
	err := ctx.ReadResource("aws:eks/accessEntry:AccessEntry", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering AccessEntry resources.
type accessEntryState struct {
	// Amazon Resource Name (ARN) of the Access Entry.
	AccessEntryArn *string `pulumi:"accessEntryArn"`
	// Name of the EKS Cluster.
	ClusterName *string `pulumi:"clusterName"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
	CreatedAt *string `pulumi:"createdAt"`
	// List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
	KubernetesGroups []string `pulumi:"kubernetesGroups"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
	ModifiedAt *string `pulumi:"modifiedAt"`
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	//
	// The following arguments are optional:
	PrincipalArn *string `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// (Optional) Key-value map of resource tags, including those inherited from the provider `defaultTags` configuration block.
	TagsAll map[string]string `pulumi:"tagsAll"`
	// Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
	Type *string `pulumi:"type"`
	// Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
	UserName *string `pulumi:"userName"`
}

type AccessEntryState struct {
	// Amazon Resource Name (ARN) of the Access Entry.
	AccessEntryArn pulumi.StringPtrInput
	// Name of the EKS Cluster.
	ClusterName pulumi.StringPtrInput
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
	CreatedAt pulumi.StringPtrInput
	// List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
	KubernetesGroups pulumi.StringArrayInput
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
	ModifiedAt pulumi.StringPtrInput
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	//
	// The following arguments are optional:
	PrincipalArn pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// (Optional) Key-value map of resource tags, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapInput
	// Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
	Type pulumi.StringPtrInput
	// Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
	UserName pulumi.StringPtrInput
}

func (AccessEntryState) ElementType() reflect.Type {
	return reflect.TypeOf((*accessEntryState)(nil)).Elem()
}

type accessEntryArgs struct {
	// Name of the EKS Cluster.
	ClusterName string `pulumi:"clusterName"`
	// List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
	KubernetesGroups []string `pulumi:"kubernetesGroups"`
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	//
	// The following arguments are optional:
	PrincipalArn string `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
	Type *string `pulumi:"type"`
	// Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
	UserName *string `pulumi:"userName"`
}

// The set of arguments for constructing a AccessEntry resource.
type AccessEntryArgs struct {
	// Name of the EKS Cluster.
	ClusterName pulumi.StringInput
	// List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
	KubernetesGroups pulumi.StringArrayInput
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	//
	// The following arguments are optional:
	PrincipalArn pulumi.StringInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
	Type pulumi.StringPtrInput
	// Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
	UserName pulumi.StringPtrInput
}

func (AccessEntryArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*accessEntryArgs)(nil)).Elem()
}

type AccessEntryInput interface {
	pulumi.Input

	ToAccessEntryOutput() AccessEntryOutput
	ToAccessEntryOutputWithContext(ctx context.Context) AccessEntryOutput
}

func (*AccessEntry) ElementType() reflect.Type {
	return reflect.TypeOf((**AccessEntry)(nil)).Elem()
}

func (i *AccessEntry) ToAccessEntryOutput() AccessEntryOutput {
	return i.ToAccessEntryOutputWithContext(context.Background())
}

func (i *AccessEntry) ToAccessEntryOutputWithContext(ctx context.Context) AccessEntryOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AccessEntryOutput)
}

// AccessEntryArrayInput is an input type that accepts AccessEntryArray and AccessEntryArrayOutput values.
// You can construct a concrete instance of `AccessEntryArrayInput` via:
//
//	AccessEntryArray{ AccessEntryArgs{...} }
type AccessEntryArrayInput interface {
	pulumi.Input

	ToAccessEntryArrayOutput() AccessEntryArrayOutput
	ToAccessEntryArrayOutputWithContext(context.Context) AccessEntryArrayOutput
}

type AccessEntryArray []AccessEntryInput

func (AccessEntryArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*AccessEntry)(nil)).Elem()
}

func (i AccessEntryArray) ToAccessEntryArrayOutput() AccessEntryArrayOutput {
	return i.ToAccessEntryArrayOutputWithContext(context.Background())
}

func (i AccessEntryArray) ToAccessEntryArrayOutputWithContext(ctx context.Context) AccessEntryArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AccessEntryArrayOutput)
}

// AccessEntryMapInput is an input type that accepts AccessEntryMap and AccessEntryMapOutput values.
// You can construct a concrete instance of `AccessEntryMapInput` via:
//
//	AccessEntryMap{ "key": AccessEntryArgs{...} }
type AccessEntryMapInput interface {
	pulumi.Input

	ToAccessEntryMapOutput() AccessEntryMapOutput
	ToAccessEntryMapOutputWithContext(context.Context) AccessEntryMapOutput
}

type AccessEntryMap map[string]AccessEntryInput

func (AccessEntryMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*AccessEntry)(nil)).Elem()
}

func (i AccessEntryMap) ToAccessEntryMapOutput() AccessEntryMapOutput {
	return i.ToAccessEntryMapOutputWithContext(context.Background())
}

func (i AccessEntryMap) ToAccessEntryMapOutputWithContext(ctx context.Context) AccessEntryMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AccessEntryMapOutput)
}

type AccessEntryOutput struct{ *pulumi.OutputState }

func (AccessEntryOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**AccessEntry)(nil)).Elem()
}

func (o AccessEntryOutput) ToAccessEntryOutput() AccessEntryOutput {
	return o
}

func (o AccessEntryOutput) ToAccessEntryOutputWithContext(ctx context.Context) AccessEntryOutput {
	return o
}

// Amazon Resource Name (ARN) of the Access Entry.
func (o AccessEntryOutput) AccessEntryArn() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.AccessEntryArn }).(pulumi.StringOutput)
}

// Name of the EKS Cluster.
func (o AccessEntryOutput) ClusterName() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.ClusterName }).(pulumi.StringOutput)
}

// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
func (o AccessEntryOutput) CreatedAt() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.CreatedAt }).(pulumi.StringOutput)
}

// List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
func (o AccessEntryOutput) KubernetesGroups() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringArrayOutput { return v.KubernetesGroups }).(pulumi.StringArrayOutput)
}

// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
func (o AccessEntryOutput) ModifiedAt() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.ModifiedAt }).(pulumi.StringOutput)
}

// The IAM Principal ARN which requires Authentication access to the EKS cluster.
//
// The following arguments are optional:
func (o AccessEntryOutput) PrincipalArn() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.PrincipalArn }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o AccessEntryOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
func (o AccessEntryOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// (Optional) Key-value map of resource tags, including those inherited from the provider `defaultTags` configuration block.
func (o AccessEntryOutput) TagsAll() pulumi.StringMapOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringMapOutput { return v.TagsAll }).(pulumi.StringMapOutput)
}

// Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
func (o AccessEntryOutput) Type() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringPtrOutput { return v.Type }).(pulumi.StringPtrOutput)
}

// Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
func (o AccessEntryOutput) UserName() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessEntry) pulumi.StringOutput { return v.UserName }).(pulumi.StringOutput)
}

type AccessEntryArrayOutput struct{ *pulumi.OutputState }

func (AccessEntryArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*AccessEntry)(nil)).Elem()
}

func (o AccessEntryArrayOutput) ToAccessEntryArrayOutput() AccessEntryArrayOutput {
	return o
}

func (o AccessEntryArrayOutput) ToAccessEntryArrayOutputWithContext(ctx context.Context) AccessEntryArrayOutput {
	return o
}

func (o AccessEntryArrayOutput) Index(i pulumi.IntInput) AccessEntryOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *AccessEntry {
		return vs[0].([]*AccessEntry)[vs[1].(int)]
	}).(AccessEntryOutput)
}

type AccessEntryMapOutput struct{ *pulumi.OutputState }

func (AccessEntryMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*AccessEntry)(nil)).Elem()
}

func (o AccessEntryMapOutput) ToAccessEntryMapOutput() AccessEntryMapOutput {
	return o
}

func (o AccessEntryMapOutput) ToAccessEntryMapOutputWithContext(ctx context.Context) AccessEntryMapOutput {
	return o
}

func (o AccessEntryMapOutput) MapIndex(k pulumi.StringInput) AccessEntryOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *AccessEntry {
		return vs[0].(map[string]*AccessEntry)[vs[1].(string)]
	}).(AccessEntryOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*AccessEntryInput)(nil)).Elem(), &AccessEntry{})
	pulumi.RegisterInputType(reflect.TypeOf((*AccessEntryArrayInput)(nil)).Elem(), AccessEntryArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*AccessEntryMapInput)(nil)).Elem(), AccessEntryMap{})
	pulumi.RegisterOutputType(AccessEntryOutput{})
	pulumi.RegisterOutputType(AccessEntryArrayOutput{})
	pulumi.RegisterOutputType(AccessEntryMapOutput{})
}
