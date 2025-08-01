// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package cloudwatch

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provides a CloudWatch Logs destination resource.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := cloudwatch.NewLogDestination(ctx, "test_destination", &cloudwatch.LogDestinationArgs{
//				Name:      pulumi.String("test_destination"),
//				RoleArn:   pulumi.Any(iamForCloudwatch.Arn),
//				TargetArn: pulumi.Any(kinesisForCloudwatch.Arn),
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
// Using `pulumi import`, import CloudWatch Logs destinations using the `name`. For example:
//
// ```sh
// $ pulumi import aws:cloudwatch/logDestination:LogDestination test_destination test_destination
// ```
type LogDestination struct {
	pulumi.CustomResourceState

	// The Amazon Resource Name (ARN) specifying the log destination.
	Arn pulumi.StringOutput `pulumi:"arn"`
	// A name for the log destination.
	Name pulumi.StringOutput `pulumi:"name"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
	RoleArn pulumi.StringOutput `pulumi:"roleArn"`
	// A map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapOutput `pulumi:"tagsAll"`
	// The ARN of the target Amazon Kinesis stream resource for the destination.
	TargetArn pulumi.StringOutput `pulumi:"targetArn"`
}

// NewLogDestination registers a new resource with the given unique name, arguments, and options.
func NewLogDestination(ctx *pulumi.Context,
	name string, args *LogDestinationArgs, opts ...pulumi.ResourceOption) (*LogDestination, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.RoleArn == nil {
		return nil, errors.New("invalid value for required argument 'RoleArn'")
	}
	if args.TargetArn == nil {
		return nil, errors.New("invalid value for required argument 'TargetArn'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource LogDestination
	err := ctx.RegisterResource("aws:cloudwatch/logDestination:LogDestination", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetLogDestination gets an existing LogDestination resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetLogDestination(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *LogDestinationState, opts ...pulumi.ResourceOption) (*LogDestination, error) {
	var resource LogDestination
	err := ctx.ReadResource("aws:cloudwatch/logDestination:LogDestination", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering LogDestination resources.
type logDestinationState struct {
	// The Amazon Resource Name (ARN) specifying the log destination.
	Arn *string `pulumi:"arn"`
	// A name for the log destination.
	Name *string `pulumi:"name"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
	RoleArn *string `pulumi:"roleArn"`
	// A map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll map[string]string `pulumi:"tagsAll"`
	// The ARN of the target Amazon Kinesis stream resource for the destination.
	TargetArn *string `pulumi:"targetArn"`
}

type LogDestinationState struct {
	// The Amazon Resource Name (ARN) specifying the log destination.
	Arn pulumi.StringPtrInput
	// A name for the log destination.
	Name pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
	RoleArn pulumi.StringPtrInput
	// A map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapInput
	// The ARN of the target Amazon Kinesis stream resource for the destination.
	TargetArn pulumi.StringPtrInput
}

func (LogDestinationState) ElementType() reflect.Type {
	return reflect.TypeOf((*logDestinationState)(nil)).Elem()
}

type logDestinationArgs struct {
	// A name for the log destination.
	Name *string `pulumi:"name"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
	RoleArn string `pulumi:"roleArn"`
	// A map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// The ARN of the target Amazon Kinesis stream resource for the destination.
	TargetArn string `pulumi:"targetArn"`
}

// The set of arguments for constructing a LogDestination resource.
type LogDestinationArgs struct {
	// A name for the log destination.
	Name pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
	RoleArn pulumi.StringInput
	// A map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// The ARN of the target Amazon Kinesis stream resource for the destination.
	TargetArn pulumi.StringInput
}

func (LogDestinationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*logDestinationArgs)(nil)).Elem()
}

type LogDestinationInput interface {
	pulumi.Input

	ToLogDestinationOutput() LogDestinationOutput
	ToLogDestinationOutputWithContext(ctx context.Context) LogDestinationOutput
}

func (*LogDestination) ElementType() reflect.Type {
	return reflect.TypeOf((**LogDestination)(nil)).Elem()
}

func (i *LogDestination) ToLogDestinationOutput() LogDestinationOutput {
	return i.ToLogDestinationOutputWithContext(context.Background())
}

func (i *LogDestination) ToLogDestinationOutputWithContext(ctx context.Context) LogDestinationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(LogDestinationOutput)
}

// LogDestinationArrayInput is an input type that accepts LogDestinationArray and LogDestinationArrayOutput values.
// You can construct a concrete instance of `LogDestinationArrayInput` via:
//
//	LogDestinationArray{ LogDestinationArgs{...} }
type LogDestinationArrayInput interface {
	pulumi.Input

	ToLogDestinationArrayOutput() LogDestinationArrayOutput
	ToLogDestinationArrayOutputWithContext(context.Context) LogDestinationArrayOutput
}

type LogDestinationArray []LogDestinationInput

func (LogDestinationArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*LogDestination)(nil)).Elem()
}

func (i LogDestinationArray) ToLogDestinationArrayOutput() LogDestinationArrayOutput {
	return i.ToLogDestinationArrayOutputWithContext(context.Background())
}

func (i LogDestinationArray) ToLogDestinationArrayOutputWithContext(ctx context.Context) LogDestinationArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(LogDestinationArrayOutput)
}

// LogDestinationMapInput is an input type that accepts LogDestinationMap and LogDestinationMapOutput values.
// You can construct a concrete instance of `LogDestinationMapInput` via:
//
//	LogDestinationMap{ "key": LogDestinationArgs{...} }
type LogDestinationMapInput interface {
	pulumi.Input

	ToLogDestinationMapOutput() LogDestinationMapOutput
	ToLogDestinationMapOutputWithContext(context.Context) LogDestinationMapOutput
}

type LogDestinationMap map[string]LogDestinationInput

func (LogDestinationMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*LogDestination)(nil)).Elem()
}

func (i LogDestinationMap) ToLogDestinationMapOutput() LogDestinationMapOutput {
	return i.ToLogDestinationMapOutputWithContext(context.Background())
}

func (i LogDestinationMap) ToLogDestinationMapOutputWithContext(ctx context.Context) LogDestinationMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(LogDestinationMapOutput)
}

type LogDestinationOutput struct{ *pulumi.OutputState }

func (LogDestinationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**LogDestination)(nil)).Elem()
}

func (o LogDestinationOutput) ToLogDestinationOutput() LogDestinationOutput {
	return o
}

func (o LogDestinationOutput) ToLogDestinationOutputWithContext(ctx context.Context) LogDestinationOutput {
	return o
}

// The Amazon Resource Name (ARN) specifying the log destination.
func (o LogDestinationOutput) Arn() pulumi.StringOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringOutput { return v.Arn }).(pulumi.StringOutput)
}

// A name for the log destination.
func (o LogDestinationOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o LogDestinationOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
func (o LogDestinationOutput) RoleArn() pulumi.StringOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringOutput { return v.RoleArn }).(pulumi.StringOutput)
}

// A map of tags to assign to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
func (o LogDestinationOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// A map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
func (o LogDestinationOutput) TagsAll() pulumi.StringMapOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringMapOutput { return v.TagsAll }).(pulumi.StringMapOutput)
}

// The ARN of the target Amazon Kinesis stream resource for the destination.
func (o LogDestinationOutput) TargetArn() pulumi.StringOutput {
	return o.ApplyT(func(v *LogDestination) pulumi.StringOutput { return v.TargetArn }).(pulumi.StringOutput)
}

type LogDestinationArrayOutput struct{ *pulumi.OutputState }

func (LogDestinationArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*LogDestination)(nil)).Elem()
}

func (o LogDestinationArrayOutput) ToLogDestinationArrayOutput() LogDestinationArrayOutput {
	return o
}

func (o LogDestinationArrayOutput) ToLogDestinationArrayOutputWithContext(ctx context.Context) LogDestinationArrayOutput {
	return o
}

func (o LogDestinationArrayOutput) Index(i pulumi.IntInput) LogDestinationOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *LogDestination {
		return vs[0].([]*LogDestination)[vs[1].(int)]
	}).(LogDestinationOutput)
}

type LogDestinationMapOutput struct{ *pulumi.OutputState }

func (LogDestinationMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*LogDestination)(nil)).Elem()
}

func (o LogDestinationMapOutput) ToLogDestinationMapOutput() LogDestinationMapOutput {
	return o
}

func (o LogDestinationMapOutput) ToLogDestinationMapOutputWithContext(ctx context.Context) LogDestinationMapOutput {
	return o
}

func (o LogDestinationMapOutput) MapIndex(k pulumi.StringInput) LogDestinationOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *LogDestination {
		return vs[0].(map[string]*LogDestination)[vs[1].(string)]
	}).(LogDestinationOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*LogDestinationInput)(nil)).Elem(), &LogDestination{})
	pulumi.RegisterInputType(reflect.TypeOf((*LogDestinationArrayInput)(nil)).Elem(), LogDestinationArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*LogDestinationMapInput)(nil)).Elem(), LogDestinationMap{})
	pulumi.RegisterOutputType(LogDestinationOutput{})
	pulumi.RegisterOutputType(LogDestinationArrayOutput{})
	pulumi.RegisterOutputType(LogDestinationMapOutput{})
}
