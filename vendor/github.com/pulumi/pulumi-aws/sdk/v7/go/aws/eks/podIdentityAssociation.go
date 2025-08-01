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

// Resource for managing an AWS EKS (Elastic Kubernetes) Pod Identity Association.
//
// Creates an EKS Pod Identity association between a service account in an Amazon EKS cluster and an IAM role with EKS Pod Identity. Use EKS Pod Identity to give temporary IAM credentials to pods and the credentials are rotated automatically.
//
// Amazon EKS Pod Identity associations provide the ability to manage credentials for your applications, similar to the way that EC2 instance profiles provide credentials to Amazon EC2 instances.
//
// If a pod uses a service account that has an association, Amazon EKS sets environment variables in the containers of the pod. The environment variables configure the Amazon Web Services SDKs, including the Command Line Interface, to use the EKS Pod Identity credentials.
//
// Pod Identity is a simpler method than IAM roles for service accounts, as this method doesn’t use OIDC identity providers. Additionally, you can configure a role for Pod Identity once, and reuse it across clusters.
//
// ## Example Usage
//
// ### Basic Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			assumeRole, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
//				Statements: []iam.GetPolicyDocumentStatement{
//					{
//						Effect: pulumi.StringRef("Allow"),
//						Principals: []iam.GetPolicyDocumentStatementPrincipal{
//							{
//								Type: "Service",
//								Identifiers: []string{
//									"pods.eks.amazonaws.com",
//								},
//							},
//						},
//						Actions: []string{
//							"sts:AssumeRole",
//							"sts:TagSession",
//						},
//					},
//				},
//			}, nil)
//			if err != nil {
//				return err
//			}
//			example, err := iam.NewRole(ctx, "example", &iam.RoleArgs{
//				Name:             pulumi.String("eks-pod-identity-example"),
//				AssumeRolePolicy: pulumi.String(assumeRole.Json),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = iam.NewRolePolicyAttachment(ctx, "example_s3", &iam.RolePolicyAttachmentArgs{
//				PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"),
//				Role:      example.Name,
//			})
//			if err != nil {
//				return err
//			}
//			_, err = eks.NewPodIdentityAssociation(ctx, "example", &eks.PodIdentityAssociationArgs{
//				ClusterName:    pulumi.Any(exampleAwsEksCluster.Name),
//				Namespace:      pulumi.String("example"),
//				ServiceAccount: pulumi.String("example-sa"),
//				RoleArn:        example.Arn,
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
// Using `pulumi import`, import EKS (Elastic Kubernetes) Pod Identity Association using the `cluster_name` and `association_id` separated by a comma (`,`). For example:
//
// ```sh
// $ pulumi import aws:eks/podIdentityAssociation:PodIdentityAssociation example example,a-12345678
// ```
type PodIdentityAssociation struct {
	pulumi.CustomResourceState

	// The Amazon Resource Name (ARN) of the association.
	AssociationArn pulumi.StringOutput `pulumi:"associationArn"`
	// The ID of the association.
	AssociationId pulumi.StringOutput `pulumi:"associationId"`
	// The name of the cluster to create the association in.
	ClusterName pulumi.StringOutput `pulumi:"clusterName"`
	// Disable the tags that are automatically added to role session by Amazon EKS.
	DisableSessionTags pulumi.BoolOutput `pulumi:"disableSessionTags"`
	// The unique identifier for this association for a target IAM role. You put this value in the trust policy of the target role, in a Condition to match the sts.ExternalId.
	ExternalId pulumi.StringOutput `pulumi:"externalId"`
	// The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
	Namespace pulumi.StringOutput `pulumi:"namespace"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
	RoleArn pulumi.StringOutput `pulumi:"roleArn"`
	// The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.
	//
	// The following arguments are optional:
	ServiceAccount pulumi.StringOutput `pulumi:"serviceAccount"`
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapOutput `pulumi:"tagsAll"`
	// The Amazon Resource Name (ARN) of the IAM role to be chained to the the IAM role specified as `roleArn`.
	TargetRoleArn pulumi.StringPtrOutput `pulumi:"targetRoleArn"`
}

// NewPodIdentityAssociation registers a new resource with the given unique name, arguments, and options.
func NewPodIdentityAssociation(ctx *pulumi.Context,
	name string, args *PodIdentityAssociationArgs, opts ...pulumi.ResourceOption) (*PodIdentityAssociation, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.ClusterName == nil {
		return nil, errors.New("invalid value for required argument 'ClusterName'")
	}
	if args.Namespace == nil {
		return nil, errors.New("invalid value for required argument 'Namespace'")
	}
	if args.RoleArn == nil {
		return nil, errors.New("invalid value for required argument 'RoleArn'")
	}
	if args.ServiceAccount == nil {
		return nil, errors.New("invalid value for required argument 'ServiceAccount'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource PodIdentityAssociation
	err := ctx.RegisterResource("aws:eks/podIdentityAssociation:PodIdentityAssociation", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetPodIdentityAssociation gets an existing PodIdentityAssociation resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetPodIdentityAssociation(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *PodIdentityAssociationState, opts ...pulumi.ResourceOption) (*PodIdentityAssociation, error) {
	var resource PodIdentityAssociation
	err := ctx.ReadResource("aws:eks/podIdentityAssociation:PodIdentityAssociation", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering PodIdentityAssociation resources.
type podIdentityAssociationState struct {
	// The Amazon Resource Name (ARN) of the association.
	AssociationArn *string `pulumi:"associationArn"`
	// The ID of the association.
	AssociationId *string `pulumi:"associationId"`
	// The name of the cluster to create the association in.
	ClusterName *string `pulumi:"clusterName"`
	// Disable the tags that are automatically added to role session by Amazon EKS.
	DisableSessionTags *bool `pulumi:"disableSessionTags"`
	// The unique identifier for this association for a target IAM role. You put this value in the trust policy of the target role, in a Condition to match the sts.ExternalId.
	ExternalId *string `pulumi:"externalId"`
	// The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
	Namespace *string `pulumi:"namespace"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
	RoleArn *string `pulumi:"roleArn"`
	// The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.
	//
	// The following arguments are optional:
	ServiceAccount *string `pulumi:"serviceAccount"`
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll map[string]string `pulumi:"tagsAll"`
	// The Amazon Resource Name (ARN) of the IAM role to be chained to the the IAM role specified as `roleArn`.
	TargetRoleArn *string `pulumi:"targetRoleArn"`
}

type PodIdentityAssociationState struct {
	// The Amazon Resource Name (ARN) of the association.
	AssociationArn pulumi.StringPtrInput
	// The ID of the association.
	AssociationId pulumi.StringPtrInput
	// The name of the cluster to create the association in.
	ClusterName pulumi.StringPtrInput
	// Disable the tags that are automatically added to role session by Amazon EKS.
	DisableSessionTags pulumi.BoolPtrInput
	// The unique identifier for this association for a target IAM role. You put this value in the trust policy of the target role, in a Condition to match the sts.ExternalId.
	ExternalId pulumi.StringPtrInput
	// The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
	Namespace pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
	RoleArn pulumi.StringPtrInput
	// The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.
	//
	// The following arguments are optional:
	ServiceAccount pulumi.StringPtrInput
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapInput
	// The Amazon Resource Name (ARN) of the IAM role to be chained to the the IAM role specified as `roleArn`.
	TargetRoleArn pulumi.StringPtrInput
}

func (PodIdentityAssociationState) ElementType() reflect.Type {
	return reflect.TypeOf((*podIdentityAssociationState)(nil)).Elem()
}

type podIdentityAssociationArgs struct {
	// The name of the cluster to create the association in.
	ClusterName string `pulumi:"clusterName"`
	// Disable the tags that are automatically added to role session by Amazon EKS.
	DisableSessionTags *bool `pulumi:"disableSessionTags"`
	// The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
	Namespace string `pulumi:"namespace"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
	RoleArn string `pulumi:"roleArn"`
	// The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.
	//
	// The following arguments are optional:
	ServiceAccount string `pulumi:"serviceAccount"`
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// The Amazon Resource Name (ARN) of the IAM role to be chained to the the IAM role specified as `roleArn`.
	TargetRoleArn *string `pulumi:"targetRoleArn"`
}

// The set of arguments for constructing a PodIdentityAssociation resource.
type PodIdentityAssociationArgs struct {
	// The name of the cluster to create the association in.
	ClusterName pulumi.StringInput
	// Disable the tags that are automatically added to role session by Amazon EKS.
	DisableSessionTags pulumi.BoolPtrInput
	// The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
	Namespace pulumi.StringInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
	RoleArn pulumi.StringInput
	// The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.
	//
	// The following arguments are optional:
	ServiceAccount pulumi.StringInput
	// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// The Amazon Resource Name (ARN) of the IAM role to be chained to the the IAM role specified as `roleArn`.
	TargetRoleArn pulumi.StringPtrInput
}

func (PodIdentityAssociationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*podIdentityAssociationArgs)(nil)).Elem()
}

type PodIdentityAssociationInput interface {
	pulumi.Input

	ToPodIdentityAssociationOutput() PodIdentityAssociationOutput
	ToPodIdentityAssociationOutputWithContext(ctx context.Context) PodIdentityAssociationOutput
}

func (*PodIdentityAssociation) ElementType() reflect.Type {
	return reflect.TypeOf((**PodIdentityAssociation)(nil)).Elem()
}

func (i *PodIdentityAssociation) ToPodIdentityAssociationOutput() PodIdentityAssociationOutput {
	return i.ToPodIdentityAssociationOutputWithContext(context.Background())
}

func (i *PodIdentityAssociation) ToPodIdentityAssociationOutputWithContext(ctx context.Context) PodIdentityAssociationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(PodIdentityAssociationOutput)
}

// PodIdentityAssociationArrayInput is an input type that accepts PodIdentityAssociationArray and PodIdentityAssociationArrayOutput values.
// You can construct a concrete instance of `PodIdentityAssociationArrayInput` via:
//
//	PodIdentityAssociationArray{ PodIdentityAssociationArgs{...} }
type PodIdentityAssociationArrayInput interface {
	pulumi.Input

	ToPodIdentityAssociationArrayOutput() PodIdentityAssociationArrayOutput
	ToPodIdentityAssociationArrayOutputWithContext(context.Context) PodIdentityAssociationArrayOutput
}

type PodIdentityAssociationArray []PodIdentityAssociationInput

func (PodIdentityAssociationArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*PodIdentityAssociation)(nil)).Elem()
}

func (i PodIdentityAssociationArray) ToPodIdentityAssociationArrayOutput() PodIdentityAssociationArrayOutput {
	return i.ToPodIdentityAssociationArrayOutputWithContext(context.Background())
}

func (i PodIdentityAssociationArray) ToPodIdentityAssociationArrayOutputWithContext(ctx context.Context) PodIdentityAssociationArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(PodIdentityAssociationArrayOutput)
}

// PodIdentityAssociationMapInput is an input type that accepts PodIdentityAssociationMap and PodIdentityAssociationMapOutput values.
// You can construct a concrete instance of `PodIdentityAssociationMapInput` via:
//
//	PodIdentityAssociationMap{ "key": PodIdentityAssociationArgs{...} }
type PodIdentityAssociationMapInput interface {
	pulumi.Input

	ToPodIdentityAssociationMapOutput() PodIdentityAssociationMapOutput
	ToPodIdentityAssociationMapOutputWithContext(context.Context) PodIdentityAssociationMapOutput
}

type PodIdentityAssociationMap map[string]PodIdentityAssociationInput

func (PodIdentityAssociationMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*PodIdentityAssociation)(nil)).Elem()
}

func (i PodIdentityAssociationMap) ToPodIdentityAssociationMapOutput() PodIdentityAssociationMapOutput {
	return i.ToPodIdentityAssociationMapOutputWithContext(context.Background())
}

func (i PodIdentityAssociationMap) ToPodIdentityAssociationMapOutputWithContext(ctx context.Context) PodIdentityAssociationMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(PodIdentityAssociationMapOutput)
}

type PodIdentityAssociationOutput struct{ *pulumi.OutputState }

func (PodIdentityAssociationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**PodIdentityAssociation)(nil)).Elem()
}

func (o PodIdentityAssociationOutput) ToPodIdentityAssociationOutput() PodIdentityAssociationOutput {
	return o
}

func (o PodIdentityAssociationOutput) ToPodIdentityAssociationOutputWithContext(ctx context.Context) PodIdentityAssociationOutput {
	return o
}

// The Amazon Resource Name (ARN) of the association.
func (o PodIdentityAssociationOutput) AssociationArn() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.AssociationArn }).(pulumi.StringOutput)
}

// The ID of the association.
func (o PodIdentityAssociationOutput) AssociationId() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.AssociationId }).(pulumi.StringOutput)
}

// The name of the cluster to create the association in.
func (o PodIdentityAssociationOutput) ClusterName() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.ClusterName }).(pulumi.StringOutput)
}

// Disable the tags that are automatically added to role session by Amazon EKS.
func (o PodIdentityAssociationOutput) DisableSessionTags() pulumi.BoolOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.BoolOutput { return v.DisableSessionTags }).(pulumi.BoolOutput)
}

// The unique identifier for this association for a target IAM role. You put this value in the trust policy of the target role, in a Condition to match the sts.ExternalId.
func (o PodIdentityAssociationOutput) ExternalId() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.ExternalId }).(pulumi.StringOutput)
}

// The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
func (o PodIdentityAssociationOutput) Namespace() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.Namespace }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o PodIdentityAssociationOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
func (o PodIdentityAssociationOutput) RoleArn() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.RoleArn }).(pulumi.StringOutput)
}

// The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.
//
// The following arguments are optional:
func (o PodIdentityAssociationOutput) ServiceAccount() pulumi.StringOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringOutput { return v.ServiceAccount }).(pulumi.StringOutput)
}

// Key-value map of resource tags. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
func (o PodIdentityAssociationOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
func (o PodIdentityAssociationOutput) TagsAll() pulumi.StringMapOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringMapOutput { return v.TagsAll }).(pulumi.StringMapOutput)
}

// The Amazon Resource Name (ARN) of the IAM role to be chained to the the IAM role specified as `roleArn`.
func (o PodIdentityAssociationOutput) TargetRoleArn() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *PodIdentityAssociation) pulumi.StringPtrOutput { return v.TargetRoleArn }).(pulumi.StringPtrOutput)
}

type PodIdentityAssociationArrayOutput struct{ *pulumi.OutputState }

func (PodIdentityAssociationArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*PodIdentityAssociation)(nil)).Elem()
}

func (o PodIdentityAssociationArrayOutput) ToPodIdentityAssociationArrayOutput() PodIdentityAssociationArrayOutput {
	return o
}

func (o PodIdentityAssociationArrayOutput) ToPodIdentityAssociationArrayOutputWithContext(ctx context.Context) PodIdentityAssociationArrayOutput {
	return o
}

func (o PodIdentityAssociationArrayOutput) Index(i pulumi.IntInput) PodIdentityAssociationOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *PodIdentityAssociation {
		return vs[0].([]*PodIdentityAssociation)[vs[1].(int)]
	}).(PodIdentityAssociationOutput)
}

type PodIdentityAssociationMapOutput struct{ *pulumi.OutputState }

func (PodIdentityAssociationMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*PodIdentityAssociation)(nil)).Elem()
}

func (o PodIdentityAssociationMapOutput) ToPodIdentityAssociationMapOutput() PodIdentityAssociationMapOutput {
	return o
}

func (o PodIdentityAssociationMapOutput) ToPodIdentityAssociationMapOutputWithContext(ctx context.Context) PodIdentityAssociationMapOutput {
	return o
}

func (o PodIdentityAssociationMapOutput) MapIndex(k pulumi.StringInput) PodIdentityAssociationOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *PodIdentityAssociation {
		return vs[0].(map[string]*PodIdentityAssociation)[vs[1].(string)]
	}).(PodIdentityAssociationOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*PodIdentityAssociationInput)(nil)).Elem(), &PodIdentityAssociation{})
	pulumi.RegisterInputType(reflect.TypeOf((*PodIdentityAssociationArrayInput)(nil)).Elem(), PodIdentityAssociationArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*PodIdentityAssociationMapInput)(nil)).Elem(), PodIdentityAssociationMap{})
	pulumi.RegisterOutputType(PodIdentityAssociationOutput{})
	pulumi.RegisterOutputType(PodIdentityAssociationArrayOutput{})
	pulumi.RegisterOutputType(PodIdentityAssociationMapOutput{})
}
