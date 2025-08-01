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

// Provides a Network Insights Analysis resource. Part of the "Reachability Analyzer" service in the AWS VPC console.
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
//			path, err := ec2.NewNetworkInsightsPath(ctx, "path", &ec2.NetworkInsightsPathArgs{
//				Source:      pulumi.Any(source.Id),
//				Destination: pulumi.Any(destination.Id),
//				Protocol:    pulumi.String("tcp"),
//			})
//			if err != nil {
//				return err
//			}
//			_, err = ec2.NewNetworkInsightsAnalysis(ctx, "analysis", &ec2.NetworkInsightsAnalysisArgs{
//				NetworkInsightsPathId: path.ID(),
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
// Using `pulumi import`, import Network Insights Analyzes using the `id`. For example:
//
// ```sh
// $ pulumi import aws:ec2/networkInsightsAnalysis:NetworkInsightsAnalysis test nia-0462085c957f11a55
// ```
type NetworkInsightsAnalysis struct {
	pulumi.CustomResourceState

	// Potential intermediate components of a feasible path. Described below.
	AlternatePathHints NetworkInsightsAnalysisAlternatePathHintArrayOutput `pulumi:"alternatePathHints"`
	// ARN of the Network Insights Analysis.
	Arn pulumi.StringOutput `pulumi:"arn"`
	// Explanation codes for an unreachable path. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Explanation.html) for details.
	Explanations NetworkInsightsAnalysisExplanationArrayOutput `pulumi:"explanations"`
	// A list of ARNs for resources the path must traverse.
	FilterInArns pulumi.StringArrayOutput `pulumi:"filterInArns"`
	// The components in the path from source to destination. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
	ForwardPathComponents NetworkInsightsAnalysisForwardPathComponentArrayOutput `pulumi:"forwardPathComponents"`
	// ID of the Network Insights Path to run an analysis on.
	//
	// The following arguments are optional:
	NetworkInsightsPathId pulumi.StringOutput `pulumi:"networkInsightsPathId"`
	// Set to `true` if the destination was reachable.
	PathFound pulumi.BoolOutput `pulumi:"pathFound"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// The components in the path from destination to source. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
	ReturnPathComponents NetworkInsightsAnalysisReturnPathComponentArrayOutput `pulumi:"returnPathComponents"`
	// The date/time the analysis was started.
	StartDate pulumi.StringOutput `pulumi:"startDate"`
	// The status of the analysis. `succeeded` means the analysis was completed, not that a path was found, for that see `pathFound`.
	Status pulumi.StringOutput `pulumi:"status"`
	// A message to provide more context when the `status` is `failed`.
	StatusMessage pulumi.StringOutput `pulumi:"statusMessage"`
	// Map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapOutput `pulumi:"tagsAll"`
	// If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
	WaitForCompletion pulumi.BoolPtrOutput `pulumi:"waitForCompletion"`
	// The warning message.
	WarningMessage pulumi.StringOutput `pulumi:"warningMessage"`
}

// NewNetworkInsightsAnalysis registers a new resource with the given unique name, arguments, and options.
func NewNetworkInsightsAnalysis(ctx *pulumi.Context,
	name string, args *NetworkInsightsAnalysisArgs, opts ...pulumi.ResourceOption) (*NetworkInsightsAnalysis, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.NetworkInsightsPathId == nil {
		return nil, errors.New("invalid value for required argument 'NetworkInsightsPathId'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource NetworkInsightsAnalysis
	err := ctx.RegisterResource("aws:ec2/networkInsightsAnalysis:NetworkInsightsAnalysis", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetNetworkInsightsAnalysis gets an existing NetworkInsightsAnalysis resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetNetworkInsightsAnalysis(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *NetworkInsightsAnalysisState, opts ...pulumi.ResourceOption) (*NetworkInsightsAnalysis, error) {
	var resource NetworkInsightsAnalysis
	err := ctx.ReadResource("aws:ec2/networkInsightsAnalysis:NetworkInsightsAnalysis", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering NetworkInsightsAnalysis resources.
type networkInsightsAnalysisState struct {
	// Potential intermediate components of a feasible path. Described below.
	AlternatePathHints []NetworkInsightsAnalysisAlternatePathHint `pulumi:"alternatePathHints"`
	// ARN of the Network Insights Analysis.
	Arn *string `pulumi:"arn"`
	// Explanation codes for an unreachable path. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Explanation.html) for details.
	Explanations []NetworkInsightsAnalysisExplanation `pulumi:"explanations"`
	// A list of ARNs for resources the path must traverse.
	FilterInArns []string `pulumi:"filterInArns"`
	// The components in the path from source to destination. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
	ForwardPathComponents []NetworkInsightsAnalysisForwardPathComponent `pulumi:"forwardPathComponents"`
	// ID of the Network Insights Path to run an analysis on.
	//
	// The following arguments are optional:
	NetworkInsightsPathId *string `pulumi:"networkInsightsPathId"`
	// Set to `true` if the destination was reachable.
	PathFound *bool `pulumi:"pathFound"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The components in the path from destination to source. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
	ReturnPathComponents []NetworkInsightsAnalysisReturnPathComponent `pulumi:"returnPathComponents"`
	// The date/time the analysis was started.
	StartDate *string `pulumi:"startDate"`
	// The status of the analysis. `succeeded` means the analysis was completed, not that a path was found, for that see `pathFound`.
	Status *string `pulumi:"status"`
	// A message to provide more context when the `status` is `failed`.
	StatusMessage *string `pulumi:"statusMessage"`
	// Map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll map[string]string `pulumi:"tagsAll"`
	// If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
	WaitForCompletion *bool `pulumi:"waitForCompletion"`
	// The warning message.
	WarningMessage *string `pulumi:"warningMessage"`
}

type NetworkInsightsAnalysisState struct {
	// Potential intermediate components of a feasible path. Described below.
	AlternatePathHints NetworkInsightsAnalysisAlternatePathHintArrayInput
	// ARN of the Network Insights Analysis.
	Arn pulumi.StringPtrInput
	// Explanation codes for an unreachable path. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Explanation.html) for details.
	Explanations NetworkInsightsAnalysisExplanationArrayInput
	// A list of ARNs for resources the path must traverse.
	FilterInArns pulumi.StringArrayInput
	// The components in the path from source to destination. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
	ForwardPathComponents NetworkInsightsAnalysisForwardPathComponentArrayInput
	// ID of the Network Insights Path to run an analysis on.
	//
	// The following arguments are optional:
	NetworkInsightsPathId pulumi.StringPtrInput
	// Set to `true` if the destination was reachable.
	PathFound pulumi.BoolPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The components in the path from destination to source. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
	ReturnPathComponents NetworkInsightsAnalysisReturnPathComponentArrayInput
	// The date/time the analysis was started.
	StartDate pulumi.StringPtrInput
	// The status of the analysis. `succeeded` means the analysis was completed, not that a path was found, for that see `pathFound`.
	Status pulumi.StringPtrInput
	// A message to provide more context when the `status` is `failed`.
	StatusMessage pulumi.StringPtrInput
	// Map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapInput
	// If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
	WaitForCompletion pulumi.BoolPtrInput
	// The warning message.
	WarningMessage pulumi.StringPtrInput
}

func (NetworkInsightsAnalysisState) ElementType() reflect.Type {
	return reflect.TypeOf((*networkInsightsAnalysisState)(nil)).Elem()
}

type networkInsightsAnalysisArgs struct {
	// A list of ARNs for resources the path must traverse.
	FilterInArns []string `pulumi:"filterInArns"`
	// ID of the Network Insights Path to run an analysis on.
	//
	// The following arguments are optional:
	NetworkInsightsPathId string `pulumi:"networkInsightsPathId"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
	WaitForCompletion *bool `pulumi:"waitForCompletion"`
}

// The set of arguments for constructing a NetworkInsightsAnalysis resource.
type NetworkInsightsAnalysisArgs struct {
	// A list of ARNs for resources the path must traverse.
	FilterInArns pulumi.StringArrayInput
	// ID of the Network Insights Path to run an analysis on.
	//
	// The following arguments are optional:
	NetworkInsightsPathId pulumi.StringInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
	WaitForCompletion pulumi.BoolPtrInput
}

func (NetworkInsightsAnalysisArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*networkInsightsAnalysisArgs)(nil)).Elem()
}

type NetworkInsightsAnalysisInput interface {
	pulumi.Input

	ToNetworkInsightsAnalysisOutput() NetworkInsightsAnalysisOutput
	ToNetworkInsightsAnalysisOutputWithContext(ctx context.Context) NetworkInsightsAnalysisOutput
}

func (*NetworkInsightsAnalysis) ElementType() reflect.Type {
	return reflect.TypeOf((**NetworkInsightsAnalysis)(nil)).Elem()
}

func (i *NetworkInsightsAnalysis) ToNetworkInsightsAnalysisOutput() NetworkInsightsAnalysisOutput {
	return i.ToNetworkInsightsAnalysisOutputWithContext(context.Background())
}

func (i *NetworkInsightsAnalysis) ToNetworkInsightsAnalysisOutputWithContext(ctx context.Context) NetworkInsightsAnalysisOutput {
	return pulumi.ToOutputWithContext(ctx, i).(NetworkInsightsAnalysisOutput)
}

// NetworkInsightsAnalysisArrayInput is an input type that accepts NetworkInsightsAnalysisArray and NetworkInsightsAnalysisArrayOutput values.
// You can construct a concrete instance of `NetworkInsightsAnalysisArrayInput` via:
//
//	NetworkInsightsAnalysisArray{ NetworkInsightsAnalysisArgs{...} }
type NetworkInsightsAnalysisArrayInput interface {
	pulumi.Input

	ToNetworkInsightsAnalysisArrayOutput() NetworkInsightsAnalysisArrayOutput
	ToNetworkInsightsAnalysisArrayOutputWithContext(context.Context) NetworkInsightsAnalysisArrayOutput
}

type NetworkInsightsAnalysisArray []NetworkInsightsAnalysisInput

func (NetworkInsightsAnalysisArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*NetworkInsightsAnalysis)(nil)).Elem()
}

func (i NetworkInsightsAnalysisArray) ToNetworkInsightsAnalysisArrayOutput() NetworkInsightsAnalysisArrayOutput {
	return i.ToNetworkInsightsAnalysisArrayOutputWithContext(context.Background())
}

func (i NetworkInsightsAnalysisArray) ToNetworkInsightsAnalysisArrayOutputWithContext(ctx context.Context) NetworkInsightsAnalysisArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(NetworkInsightsAnalysisArrayOutput)
}

// NetworkInsightsAnalysisMapInput is an input type that accepts NetworkInsightsAnalysisMap and NetworkInsightsAnalysisMapOutput values.
// You can construct a concrete instance of `NetworkInsightsAnalysisMapInput` via:
//
//	NetworkInsightsAnalysisMap{ "key": NetworkInsightsAnalysisArgs{...} }
type NetworkInsightsAnalysisMapInput interface {
	pulumi.Input

	ToNetworkInsightsAnalysisMapOutput() NetworkInsightsAnalysisMapOutput
	ToNetworkInsightsAnalysisMapOutputWithContext(context.Context) NetworkInsightsAnalysisMapOutput
}

type NetworkInsightsAnalysisMap map[string]NetworkInsightsAnalysisInput

func (NetworkInsightsAnalysisMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*NetworkInsightsAnalysis)(nil)).Elem()
}

func (i NetworkInsightsAnalysisMap) ToNetworkInsightsAnalysisMapOutput() NetworkInsightsAnalysisMapOutput {
	return i.ToNetworkInsightsAnalysisMapOutputWithContext(context.Background())
}

func (i NetworkInsightsAnalysisMap) ToNetworkInsightsAnalysisMapOutputWithContext(ctx context.Context) NetworkInsightsAnalysisMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(NetworkInsightsAnalysisMapOutput)
}

type NetworkInsightsAnalysisOutput struct{ *pulumi.OutputState }

func (NetworkInsightsAnalysisOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**NetworkInsightsAnalysis)(nil)).Elem()
}

func (o NetworkInsightsAnalysisOutput) ToNetworkInsightsAnalysisOutput() NetworkInsightsAnalysisOutput {
	return o
}

func (o NetworkInsightsAnalysisOutput) ToNetworkInsightsAnalysisOutputWithContext(ctx context.Context) NetworkInsightsAnalysisOutput {
	return o
}

// Potential intermediate components of a feasible path. Described below.
func (o NetworkInsightsAnalysisOutput) AlternatePathHints() NetworkInsightsAnalysisAlternatePathHintArrayOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) NetworkInsightsAnalysisAlternatePathHintArrayOutput {
		return v.AlternatePathHints
	}).(NetworkInsightsAnalysisAlternatePathHintArrayOutput)
}

// ARN of the Network Insights Analysis.
func (o NetworkInsightsAnalysisOutput) Arn() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.Arn }).(pulumi.StringOutput)
}

// Explanation codes for an unreachable path. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Explanation.html) for details.
func (o NetworkInsightsAnalysisOutput) Explanations() NetworkInsightsAnalysisExplanationArrayOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) NetworkInsightsAnalysisExplanationArrayOutput { return v.Explanations }).(NetworkInsightsAnalysisExplanationArrayOutput)
}

// A list of ARNs for resources the path must traverse.
func (o NetworkInsightsAnalysisOutput) FilterInArns() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringArrayOutput { return v.FilterInArns }).(pulumi.StringArrayOutput)
}

// The components in the path from source to destination. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
func (o NetworkInsightsAnalysisOutput) ForwardPathComponents() NetworkInsightsAnalysisForwardPathComponentArrayOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) NetworkInsightsAnalysisForwardPathComponentArrayOutput {
		return v.ForwardPathComponents
	}).(NetworkInsightsAnalysisForwardPathComponentArrayOutput)
}

// ID of the Network Insights Path to run an analysis on.
//
// The following arguments are optional:
func (o NetworkInsightsAnalysisOutput) NetworkInsightsPathId() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.NetworkInsightsPathId }).(pulumi.StringOutput)
}

// Set to `true` if the destination was reachable.
func (o NetworkInsightsAnalysisOutput) PathFound() pulumi.BoolOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.BoolOutput { return v.PathFound }).(pulumi.BoolOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o NetworkInsightsAnalysisOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// The components in the path from destination to source. See the [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_PathComponent.html) for details.
func (o NetworkInsightsAnalysisOutput) ReturnPathComponents() NetworkInsightsAnalysisReturnPathComponentArrayOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) NetworkInsightsAnalysisReturnPathComponentArrayOutput {
		return v.ReturnPathComponents
	}).(NetworkInsightsAnalysisReturnPathComponentArrayOutput)
}

// The date/time the analysis was started.
func (o NetworkInsightsAnalysisOutput) StartDate() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.StartDate }).(pulumi.StringOutput)
}

// The status of the analysis. `succeeded` means the analysis was completed, not that a path was found, for that see `pathFound`.
func (o NetworkInsightsAnalysisOutput) Status() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.Status }).(pulumi.StringOutput)
}

// A message to provide more context when the `status` is `failed`.
func (o NetworkInsightsAnalysisOutput) StatusMessage() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.StatusMessage }).(pulumi.StringOutput)
}

// Map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
func (o NetworkInsightsAnalysisOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
func (o NetworkInsightsAnalysisOutput) TagsAll() pulumi.StringMapOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringMapOutput { return v.TagsAll }).(pulumi.StringMapOutput)
}

// If enabled, the resource will wait for the Network Insights Analysis status to change to `succeeded` or `failed`. Setting this to `false` will skip the process. Default: `true`.
func (o NetworkInsightsAnalysisOutput) WaitForCompletion() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.BoolPtrOutput { return v.WaitForCompletion }).(pulumi.BoolPtrOutput)
}

// The warning message.
func (o NetworkInsightsAnalysisOutput) WarningMessage() pulumi.StringOutput {
	return o.ApplyT(func(v *NetworkInsightsAnalysis) pulumi.StringOutput { return v.WarningMessage }).(pulumi.StringOutput)
}

type NetworkInsightsAnalysisArrayOutput struct{ *pulumi.OutputState }

func (NetworkInsightsAnalysisArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*NetworkInsightsAnalysis)(nil)).Elem()
}

func (o NetworkInsightsAnalysisArrayOutput) ToNetworkInsightsAnalysisArrayOutput() NetworkInsightsAnalysisArrayOutput {
	return o
}

func (o NetworkInsightsAnalysisArrayOutput) ToNetworkInsightsAnalysisArrayOutputWithContext(ctx context.Context) NetworkInsightsAnalysisArrayOutput {
	return o
}

func (o NetworkInsightsAnalysisArrayOutput) Index(i pulumi.IntInput) NetworkInsightsAnalysisOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *NetworkInsightsAnalysis {
		return vs[0].([]*NetworkInsightsAnalysis)[vs[1].(int)]
	}).(NetworkInsightsAnalysisOutput)
}

type NetworkInsightsAnalysisMapOutput struct{ *pulumi.OutputState }

func (NetworkInsightsAnalysisMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*NetworkInsightsAnalysis)(nil)).Elem()
}

func (o NetworkInsightsAnalysisMapOutput) ToNetworkInsightsAnalysisMapOutput() NetworkInsightsAnalysisMapOutput {
	return o
}

func (o NetworkInsightsAnalysisMapOutput) ToNetworkInsightsAnalysisMapOutputWithContext(ctx context.Context) NetworkInsightsAnalysisMapOutput {
	return o
}

func (o NetworkInsightsAnalysisMapOutput) MapIndex(k pulumi.StringInput) NetworkInsightsAnalysisOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *NetworkInsightsAnalysis {
		return vs[0].(map[string]*NetworkInsightsAnalysis)[vs[1].(string)]
	}).(NetworkInsightsAnalysisOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*NetworkInsightsAnalysisInput)(nil)).Elem(), &NetworkInsightsAnalysis{})
	pulumi.RegisterInputType(reflect.TypeOf((*NetworkInsightsAnalysisArrayInput)(nil)).Elem(), NetworkInsightsAnalysisArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*NetworkInsightsAnalysisMapInput)(nil)).Elem(), NetworkInsightsAnalysisMap{})
	pulumi.RegisterOutputType(NetworkInsightsAnalysisOutput{})
	pulumi.RegisterOutputType(NetworkInsightsAnalysisArrayOutput{})
	pulumi.RegisterOutputType(NetworkInsightsAnalysisMapOutput{})
}
