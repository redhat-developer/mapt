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

// Access Entry Policy Association for an EKS Cluster.
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
//			_, err := eks.NewAccessPolicyAssociation(ctx, "example", &eks.AccessPolicyAssociationArgs{
//				ClusterName:  pulumi.Any(exampleAwsEksCluster.Name),
//				PolicyArn:    pulumi.String("arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"),
//				PrincipalArn: pulumi.Any(exampleAwsIamUser.Arn),
//				AccessScope: &eks.AccessPolicyAssociationAccessScopeArgs{
//					Type: pulumi.String("namespace"),
//					Namespaces: pulumi.StringArray{
//						pulumi.String("example-namespace"),
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
// Using `pulumi import`, import EKS access entry using the `cluster_name` `principal_arn` and `policy_arn` separated by an octothorp (`#`). For example:
//
// ```sh
// $ pulumi import aws:eks/accessPolicyAssociation:AccessPolicyAssociation my_eks_access_entry my_cluster_name#my_principal_arn#my_policy_arn
// ```
type AccessPolicyAssociation struct {
	pulumi.CustomResourceState

	// The configuration block to determine the scope of the access. See `accessScope` Block below.
	AccessScope AccessPolicyAssociationAccessScopeOutput `pulumi:"accessScope"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was associated.
	AssociatedAt pulumi.StringOutput `pulumi:"associatedAt"`
	// Name of the EKS Cluster.
	ClusterName pulumi.StringOutput `pulumi:"clusterName"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was updated.
	ModifiedAt pulumi.StringOutput `pulumi:"modifiedAt"`
	// The ARN of the access policy that you're associating.
	PolicyArn pulumi.StringOutput `pulumi:"policyArn"`
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	PrincipalArn pulumi.StringOutput `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
}

// NewAccessPolicyAssociation registers a new resource with the given unique name, arguments, and options.
func NewAccessPolicyAssociation(ctx *pulumi.Context,
	name string, args *AccessPolicyAssociationArgs, opts ...pulumi.ResourceOption) (*AccessPolicyAssociation, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.AccessScope == nil {
		return nil, errors.New("invalid value for required argument 'AccessScope'")
	}
	if args.ClusterName == nil {
		return nil, errors.New("invalid value for required argument 'ClusterName'")
	}
	if args.PolicyArn == nil {
		return nil, errors.New("invalid value for required argument 'PolicyArn'")
	}
	if args.PrincipalArn == nil {
		return nil, errors.New("invalid value for required argument 'PrincipalArn'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource AccessPolicyAssociation
	err := ctx.RegisterResource("aws:eks/accessPolicyAssociation:AccessPolicyAssociation", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetAccessPolicyAssociation gets an existing AccessPolicyAssociation resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetAccessPolicyAssociation(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *AccessPolicyAssociationState, opts ...pulumi.ResourceOption) (*AccessPolicyAssociation, error) {
	var resource AccessPolicyAssociation
	err := ctx.ReadResource("aws:eks/accessPolicyAssociation:AccessPolicyAssociation", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering AccessPolicyAssociation resources.
type accessPolicyAssociationState struct {
	// The configuration block to determine the scope of the access. See `accessScope` Block below.
	AccessScope *AccessPolicyAssociationAccessScope `pulumi:"accessScope"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was associated.
	AssociatedAt *string `pulumi:"associatedAt"`
	// Name of the EKS Cluster.
	ClusterName *string `pulumi:"clusterName"`
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was updated.
	ModifiedAt *string `pulumi:"modifiedAt"`
	// The ARN of the access policy that you're associating.
	PolicyArn *string `pulumi:"policyArn"`
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	PrincipalArn *string `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

type AccessPolicyAssociationState struct {
	// The configuration block to determine the scope of the access. See `accessScope` Block below.
	AccessScope AccessPolicyAssociationAccessScopePtrInput
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was associated.
	AssociatedAt pulumi.StringPtrInput
	// Name of the EKS Cluster.
	ClusterName pulumi.StringPtrInput
	// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was updated.
	ModifiedAt pulumi.StringPtrInput
	// The ARN of the access policy that you're associating.
	PolicyArn pulumi.StringPtrInput
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	PrincipalArn pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
}

func (AccessPolicyAssociationState) ElementType() reflect.Type {
	return reflect.TypeOf((*accessPolicyAssociationState)(nil)).Elem()
}

type accessPolicyAssociationArgs struct {
	// The configuration block to determine the scope of the access. See `accessScope` Block below.
	AccessScope AccessPolicyAssociationAccessScope `pulumi:"accessScope"`
	// Name of the EKS Cluster.
	ClusterName string `pulumi:"clusterName"`
	// The ARN of the access policy that you're associating.
	PolicyArn string `pulumi:"policyArn"`
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	PrincipalArn string `pulumi:"principalArn"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

// The set of arguments for constructing a AccessPolicyAssociation resource.
type AccessPolicyAssociationArgs struct {
	// The configuration block to determine the scope of the access. See `accessScope` Block below.
	AccessScope AccessPolicyAssociationAccessScopeInput
	// Name of the EKS Cluster.
	ClusterName pulumi.StringInput
	// The ARN of the access policy that you're associating.
	PolicyArn pulumi.StringInput
	// The IAM Principal ARN which requires Authentication access to the EKS cluster.
	PrincipalArn pulumi.StringInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
}

func (AccessPolicyAssociationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*accessPolicyAssociationArgs)(nil)).Elem()
}

type AccessPolicyAssociationInput interface {
	pulumi.Input

	ToAccessPolicyAssociationOutput() AccessPolicyAssociationOutput
	ToAccessPolicyAssociationOutputWithContext(ctx context.Context) AccessPolicyAssociationOutput
}

func (*AccessPolicyAssociation) ElementType() reflect.Type {
	return reflect.TypeOf((**AccessPolicyAssociation)(nil)).Elem()
}

func (i *AccessPolicyAssociation) ToAccessPolicyAssociationOutput() AccessPolicyAssociationOutput {
	return i.ToAccessPolicyAssociationOutputWithContext(context.Background())
}

func (i *AccessPolicyAssociation) ToAccessPolicyAssociationOutputWithContext(ctx context.Context) AccessPolicyAssociationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AccessPolicyAssociationOutput)
}

// AccessPolicyAssociationArrayInput is an input type that accepts AccessPolicyAssociationArray and AccessPolicyAssociationArrayOutput values.
// You can construct a concrete instance of `AccessPolicyAssociationArrayInput` via:
//
//	AccessPolicyAssociationArray{ AccessPolicyAssociationArgs{...} }
type AccessPolicyAssociationArrayInput interface {
	pulumi.Input

	ToAccessPolicyAssociationArrayOutput() AccessPolicyAssociationArrayOutput
	ToAccessPolicyAssociationArrayOutputWithContext(context.Context) AccessPolicyAssociationArrayOutput
}

type AccessPolicyAssociationArray []AccessPolicyAssociationInput

func (AccessPolicyAssociationArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*AccessPolicyAssociation)(nil)).Elem()
}

func (i AccessPolicyAssociationArray) ToAccessPolicyAssociationArrayOutput() AccessPolicyAssociationArrayOutput {
	return i.ToAccessPolicyAssociationArrayOutputWithContext(context.Background())
}

func (i AccessPolicyAssociationArray) ToAccessPolicyAssociationArrayOutputWithContext(ctx context.Context) AccessPolicyAssociationArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AccessPolicyAssociationArrayOutput)
}

// AccessPolicyAssociationMapInput is an input type that accepts AccessPolicyAssociationMap and AccessPolicyAssociationMapOutput values.
// You can construct a concrete instance of `AccessPolicyAssociationMapInput` via:
//
//	AccessPolicyAssociationMap{ "key": AccessPolicyAssociationArgs{...} }
type AccessPolicyAssociationMapInput interface {
	pulumi.Input

	ToAccessPolicyAssociationMapOutput() AccessPolicyAssociationMapOutput
	ToAccessPolicyAssociationMapOutputWithContext(context.Context) AccessPolicyAssociationMapOutput
}

type AccessPolicyAssociationMap map[string]AccessPolicyAssociationInput

func (AccessPolicyAssociationMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*AccessPolicyAssociation)(nil)).Elem()
}

func (i AccessPolicyAssociationMap) ToAccessPolicyAssociationMapOutput() AccessPolicyAssociationMapOutput {
	return i.ToAccessPolicyAssociationMapOutputWithContext(context.Background())
}

func (i AccessPolicyAssociationMap) ToAccessPolicyAssociationMapOutputWithContext(ctx context.Context) AccessPolicyAssociationMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(AccessPolicyAssociationMapOutput)
}

type AccessPolicyAssociationOutput struct{ *pulumi.OutputState }

func (AccessPolicyAssociationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**AccessPolicyAssociation)(nil)).Elem()
}

func (o AccessPolicyAssociationOutput) ToAccessPolicyAssociationOutput() AccessPolicyAssociationOutput {
	return o
}

func (o AccessPolicyAssociationOutput) ToAccessPolicyAssociationOutputWithContext(ctx context.Context) AccessPolicyAssociationOutput {
	return o
}

// The configuration block to determine the scope of the access. See `accessScope` Block below.
func (o AccessPolicyAssociationOutput) AccessScope() AccessPolicyAssociationAccessScopeOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) AccessPolicyAssociationAccessScopeOutput { return v.AccessScope }).(AccessPolicyAssociationAccessScopeOutput)
}

// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was associated.
func (o AccessPolicyAssociationOutput) AssociatedAt() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) pulumi.StringOutput { return v.AssociatedAt }).(pulumi.StringOutput)
}

// Name of the EKS Cluster.
func (o AccessPolicyAssociationOutput) ClusterName() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) pulumi.StringOutput { return v.ClusterName }).(pulumi.StringOutput)
}

// Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was updated.
func (o AccessPolicyAssociationOutput) ModifiedAt() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) pulumi.StringOutput { return v.ModifiedAt }).(pulumi.StringOutput)
}

// The ARN of the access policy that you're associating.
func (o AccessPolicyAssociationOutput) PolicyArn() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) pulumi.StringOutput { return v.PolicyArn }).(pulumi.StringOutput)
}

// The IAM Principal ARN which requires Authentication access to the EKS cluster.
func (o AccessPolicyAssociationOutput) PrincipalArn() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) pulumi.StringOutput { return v.PrincipalArn }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o AccessPolicyAssociationOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *AccessPolicyAssociation) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

type AccessPolicyAssociationArrayOutput struct{ *pulumi.OutputState }

func (AccessPolicyAssociationArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*AccessPolicyAssociation)(nil)).Elem()
}

func (o AccessPolicyAssociationArrayOutput) ToAccessPolicyAssociationArrayOutput() AccessPolicyAssociationArrayOutput {
	return o
}

func (o AccessPolicyAssociationArrayOutput) ToAccessPolicyAssociationArrayOutputWithContext(ctx context.Context) AccessPolicyAssociationArrayOutput {
	return o
}

func (o AccessPolicyAssociationArrayOutput) Index(i pulumi.IntInput) AccessPolicyAssociationOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *AccessPolicyAssociation {
		return vs[0].([]*AccessPolicyAssociation)[vs[1].(int)]
	}).(AccessPolicyAssociationOutput)
}

type AccessPolicyAssociationMapOutput struct{ *pulumi.OutputState }

func (AccessPolicyAssociationMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*AccessPolicyAssociation)(nil)).Elem()
}

func (o AccessPolicyAssociationMapOutput) ToAccessPolicyAssociationMapOutput() AccessPolicyAssociationMapOutput {
	return o
}

func (o AccessPolicyAssociationMapOutput) ToAccessPolicyAssociationMapOutputWithContext(ctx context.Context) AccessPolicyAssociationMapOutput {
	return o
}

func (o AccessPolicyAssociationMapOutput) MapIndex(k pulumi.StringInput) AccessPolicyAssociationOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *AccessPolicyAssociation {
		return vs[0].(map[string]*AccessPolicyAssociation)[vs[1].(string)]
	}).(AccessPolicyAssociationOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*AccessPolicyAssociationInput)(nil)).Elem(), &AccessPolicyAssociation{})
	pulumi.RegisterInputType(reflect.TypeOf((*AccessPolicyAssociationArrayInput)(nil)).Elem(), AccessPolicyAssociationArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*AccessPolicyAssociationMapInput)(nil)).Elem(), AccessPolicyAssociationMap{})
	pulumi.RegisterOutputType(AccessPolicyAssociationOutput{})
	pulumi.RegisterOutputType(AccessPolicyAssociationArrayOutput{})
	pulumi.RegisterOutputType(AccessPolicyAssociationMapOutput{})
}
