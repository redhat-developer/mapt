// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package cloudwatch

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provides an EventBridge event bus resource.
//
// > **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.
//
// ## Example Usage
//
// ### Basic Usages
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
//			_, err := cloudwatch.NewEventBus(ctx, "messenger", &cloudwatch.EventBusArgs{
//				Name: pulumi.String("chat-messages"),
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
//			examplepartner, err := cloudwatch.GetEventSource(ctx, &cloudwatch.GetEventSourceArgs{
//				NamePrefix: pulumi.StringRef("aws.partner/examplepartner.com"),
//			}, nil)
//			if err != nil {
//				return err
//			}
//			_, err = cloudwatch.NewEventBus(ctx, "examplepartner", &cloudwatch.EventBusArgs{
//				Name:            pulumi.String(examplepartner.Name),
//				Description:     pulumi.String("Event bus for example partner events"),
//				EventSourceName: pulumi.String(examplepartner.Name),
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
// ### Logging to CloudWatch Logs, S3, and Data Firehose
//
// See [Configuring logs for Amazon EventBridge event buses](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-bus-logs.html) for more details.
//
// #### Required Resources
//
// * EventBridge Event Bus with `logConfig` configured
// * Log destinations:
//
//   - CloudWatch Logs log group
//   - S3 bucket
//   - Data Firehose delivery stream
//
// * Resource-based policy or tagging for the service-linked role:
//
//   - CloudWatch Logs log group - `cloudwatch.LogResourcePolicy` to allow `delivery.logs.amazonaws.com` to put logs into the log group
//   - S3 bucket - `s3.BucketPolicy` to allow `delivery.logs.amazonaws.com` to put logs into the bucket
//   - Data Firehose delivery stream - tagging the delivery stream with `LogDeliveryEnabled = "true"` to allow the service-linked role `AWSServiceRoleForLogDelivery` to deliver logs
//
// * CloudWatch Logs Delivery:
//
//   - `cloudwatch.LogDeliverySource` for each log type (INFO, ERROR, TRACE)
//   - `cloudwatch.LogDeliveryDestination` for the log destination (S3 bucket, CloudWatch Logs log group, or Data Firehose delivery stream)
//   - `cloudwatch.LogDelivery` to link each log type’s delivery source to the delivery destination
//
// ### Example Usage
//
// The following example demonstrates how to set up logging for an EventBridge event bus to all three destinations: CloudWatch Logs, S3, and Data Firehose.
//
// ```go
// package main
//
// import (
//
//	"fmt"
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
// func main() {
// pulumi.Run(func(ctx *pulumi.Context) error {
// current, err := aws.GetCallerIdentity(ctx, &aws.GetCallerIdentityArgs{
// }, nil);
// if err != nil {
// return err
// }
// example, err := cloudwatch.NewEventBus(ctx, "example", &cloudwatch.EventBusArgs{
// Name: pulumi.String("example-event-bus"),
// LogConfig: &cloudwatch.EventBusLogConfigArgs{
// IncludeDetail: pulumi.String("FULL"),
// Level: pulumi.String("TRACE"),
// },
// })
// if err != nil {
// return err
// }
// // CloudWatch Log Delivery Sources for INFO, ERROR, and TRACE logs
// infoLogs, err := cloudwatch.NewLogDeliverySource(ctx, "info_logs", &cloudwatch.LogDeliverySourceArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("EventBusSource-%v-INFO_LOGS", name), nil
// }).(pulumi.StringOutput),
// LogType: pulumi.String("INFO_LOGS"),
// ResourceArn: example.Arn,
// })
// if err != nil {
// return err
// }
// errorLogs, err := cloudwatch.NewLogDeliverySource(ctx, "error_logs", &cloudwatch.LogDeliverySourceArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("EventBusSource-%v-ERROR_LOGS", name), nil
// }).(pulumi.StringOutput),
// LogType: pulumi.String("ERROR_LOGS"),
// ResourceArn: example.Arn,
// })
// if err != nil {
// return err
// }
// traceLogs, err := cloudwatch.NewLogDeliverySource(ctx, "trace_logs", &cloudwatch.LogDeliverySourceArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("EventBusSource-%v-TRACE_LOGS", name), nil
// }).(pulumi.StringOutput),
// LogType: pulumi.String("TRACE_LOGS"),
// ResourceArn: example.Arn,
// })
// if err != nil {
// return err
// }
// // Logging to S3 Bucket
// exampleBucket, err := s3.NewBucket(ctx, "example", &s3.BucketArgs{
// Bucket: pulumi.String("example-event-bus-logs"),
// })
// if err != nil {
// return err
// }
// bucket := pulumi.All(exampleBucket.Arn,infoLogs.Arn,errorLogs.Arn,traceLogs.Arn).ApplyT(func(_args []interface{}) (iam.GetPolicyDocumentResult, error) {
// exampleBucketArn := _args[0].(string)
// infoLogsArn := _args[1].(string)
// errorLogsArn := _args[2].(string)
// traceLogsArn := _args[3].(string)
// return iam.GetPolicyDocumentResult(iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
// Statements: []iam.GetPolicyDocumentStatement{
// {
// Effect: pulumi.StringRef(pulumi.String(pulumi.StringRef("Allow"))),
// Principals: []iam.GetPolicyDocumentStatementPrincipal{
// {
// Type: "Service",
// Identifiers: []string{
// "delivery.logs.amazonaws.com",
// },
// },
// },
// Actions: []string{
// "s3:PutObject",
// },
// Resources: []string{
// fmt.Sprintf("%v/AWSLogs/%v/EventBusLogs/*", exampleBucketArn, current.AccountId),
// },
// Conditions: []iam.GetPolicyDocumentStatementCondition{
// {
// Test: "StringEquals",
// Variable: "s3:x-amz-acl",
// Values: []string{
// "bucket-owner-full-control",
// },
// },
// {
// Test: "StringEquals",
// Variable: "aws:SourceAccount",
// Values: interface{}{
// current.AccountId,
// },
// },
// {
// Test: "ArnLike",
// Variable: "aws:SourceArn",
// Values: []string{
// infoLogsArn,
// errorLogsArn,
// traceLogsArn,
// },
// },
// },
// },
// },
// }, nil)), nil
// }).(iam.GetPolicyDocumentResultOutput)
// _, err = s3.NewBucketPolicy(ctx, "example", &s3.BucketPolicyArgs{
// Bucket: exampleBucket.Bucket,
// Policy: pulumi.String(bucket.Json),
// })
// if err != nil {
// return err
// }
// s3, err := cloudwatch.NewLogDeliveryDestination(ctx, "s3", &cloudwatch.LogDeliveryDestinationArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("EventsDeliveryDestination-%v-S3", name), nil
// }).(pulumi.StringOutput),
// DeliveryDestinationConfiguration: &cloudwatch.LogDeliveryDestinationDeliveryDestinationConfigurationArgs{
// DestinationResourceArn: exampleBucket.Arn,
// },
// })
// if err != nil {
// return err
// }
// s3InfoLogs, err := cloudwatch.NewLogDelivery(ctx, "s3_info_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: s3.Arn,
// DeliverySourceName: infoLogs.Name,
// })
// if err != nil {
// return err
// }
// s3ErrorLogs, err := cloudwatch.NewLogDelivery(ctx, "s3_error_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: s3.Arn,
// DeliverySourceName: errorLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// s3InfoLogs,
// }))
// if err != nil {
// return err
// }
// s3TraceLogs, err := cloudwatch.NewLogDelivery(ctx, "s3_trace_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: s3.Arn,
// DeliverySourceName: traceLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// s3ErrorLogs,
// }))
// if err != nil {
// return err
// }
// // Logging to CloudWatch Log Group
// eventBusLogs, err := cloudwatch.NewLogGroup(ctx, "event_bus_logs", &cloudwatch.LogGroupArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("/aws/vendedlogs/events/event-bus/%v", name), nil
// }).(pulumi.StringOutput),
// })
// if err != nil {
// return err
// }
// cwlogs := pulumi.All(eventBusLogs.Arn,infoLogs.Arn,errorLogs.Arn,traceLogs.Arn).ApplyT(func(_args []interface{}) (iam.GetPolicyDocumentResult, error) {
// eventBusLogsArn := _args[0].(string)
// infoLogsArn := _args[1].(string)
// errorLogsArn := _args[2].(string)
// traceLogsArn := _args[3].(string)
// return iam.GetPolicyDocumentResult(iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
// Statements: []iam.GetPolicyDocumentStatement{
// {
// Effect: pulumi.StringRef(pulumi.String(pulumi.StringRef("Allow"))),
// Principals: []iam.GetPolicyDocumentStatementPrincipal{
// {
// Type: "Service",
// Identifiers: []string{
// "delivery.logs.amazonaws.com",
// },
// },
// },
// Actions: []string{
// "logs:CreateLogStream",
// "logs:PutLogEvents",
// },
// Resources: []string{
// fmt.Sprintf("%v:log-stream:*", eventBusLogsArn),
// },
// Conditions: []iam.GetPolicyDocumentStatementCondition{
// {
// Test: "StringEquals",
// Variable: "aws:SourceAccount",
// Values: interface{}{
// current.AccountId,
// },
// },
// {
// Test: "ArnLike",
// Variable: "aws:SourceArn",
// Values: []string{
// infoLogsArn,
// errorLogsArn,
// traceLogsArn,
// },
// },
// },
// },
// },
// }, nil)), nil
// }).(iam.GetPolicyDocumentResultOutput)
// _, err = cloudwatch.NewLogResourcePolicy(ctx, "example", &cloudwatch.LogResourcePolicyArgs{
// PolicyDocument: pulumi.String(cwlogs.Json),
// PolicyName: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("AWSLogDeliveryWrite-%v", name), nil
// }).(pulumi.StringOutput),
// })
// if err != nil {
// return err
// }
// cwlogsLogDeliveryDestination, err := cloudwatch.NewLogDeliveryDestination(ctx, "cwlogs", &cloudwatch.LogDeliveryDestinationArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("EventsDeliveryDestination-%v-CWLogs", name), nil
// }).(pulumi.StringOutput),
// DeliveryDestinationConfiguration: &cloudwatch.LogDeliveryDestinationDeliveryDestinationConfigurationArgs{
// DestinationResourceArn: eventBusLogs.Arn,
// },
// })
// if err != nil {
// return err
// }
// cwlogsInfoLogs, err := cloudwatch.NewLogDelivery(ctx, "cwlogs_info_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: cwlogsLogDeliveryDestination.Arn,
// DeliverySourceName: infoLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// s3InfoLogs,
// }))
// if err != nil {
// return err
// }
// cwlogsErrorLogs, err := cloudwatch.NewLogDelivery(ctx, "cwlogs_error_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: cwlogsLogDeliveryDestination.Arn,
// DeliverySourceName: errorLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// s3ErrorLogs,
// cwlogsInfoLogs,
// }))
// if err != nil {
// return err
// }
// cwlogsTraceLogs, err := cloudwatch.NewLogDelivery(ctx, "cwlogs_trace_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: cwlogsLogDeliveryDestination.Arn,
// DeliverySourceName: traceLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// s3TraceLogs,
// cwlogsErrorLogs,
// }))
// if err != nil {
// return err
// }
// // Logging to Data Firehose
// cloudfrontLogs, err := kinesis.NewFirehoseDeliveryStream(ctx, "cloudfront_logs", &kinesis.FirehoseDeliveryStreamArgs{
// Tags: pulumi.StringMap{
// "LogDeliveryEnabled": pulumi.String("true"),
// },
// })
// if err != nil {
// return err
// }
// firehose, err := cloudwatch.NewLogDeliveryDestination(ctx, "firehose", &cloudwatch.LogDeliveryDestinationArgs{
// Name: example.Name.ApplyT(func(name string) (string, error) {
// return fmt.Sprintf("EventsDeliveryDestination-%v-Firehose", name), nil
// }).(pulumi.StringOutput),
// DeliveryDestinationConfiguration: &cloudwatch.LogDeliveryDestinationDeliveryDestinationConfigurationArgs{
// DestinationResourceArn: cloudfrontLogs.Arn,
// },
// })
// if err != nil {
// return err
// }
// firehoseInfoLogs, err := cloudwatch.NewLogDelivery(ctx, "firehose_info_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: firehose.Arn,
// DeliverySourceName: infoLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// cwlogsInfoLogs,
// }))
// if err != nil {
// return err
// }
// firehoseErrorLogs, err := cloudwatch.NewLogDelivery(ctx, "firehose_error_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: firehose.Arn,
// DeliverySourceName: errorLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// cwlogsErrorLogs,
// firehoseInfoLogs,
// }))
// if err != nil {
// return err
// }
// _, err = cloudwatch.NewLogDelivery(ctx, "firehose_trace_logs", &cloudwatch.LogDeliveryArgs{
// DeliveryDestinationArn: firehose.Arn,
// DeliverySourceName: traceLogs.Name,
// }, pulumi.DependsOn([]pulumi.Resource{
// cwlogsTraceLogs,
// firehoseErrorLogs,
// }))
// if err != nil {
// return err
// }
// return nil
// })
// }
// ```
//
// ## Import
//
// Using `pulumi import`, import EventBridge event buses using the name of the event bus (which can also be a partner event source name). For example:
//
// ```sh
// $ pulumi import aws:cloudwatch/eventBus:EventBus messenger chat-messages
// ```
type EventBus struct {
	pulumi.CustomResourceState

	// ARN of the event bus.
	Arn pulumi.StringOutput `pulumi:"arn"`
	// Configuration details of the Amazon SQS queue for EventBridge to use as a dead-letter queue (DLQ). This block supports the following arguments:
	DeadLetterConfig EventBusDeadLetterConfigPtrOutput `pulumi:"deadLetterConfig"`
	// Event bus description.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// Partner event source that the new event bus will be matched with. Must match `name`.
	EventSourceName pulumi.StringPtrOutput `pulumi:"eventSourceName"`
	// Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
	KmsKeyIdentifier pulumi.StringPtrOutput `pulumi:"kmsKeyIdentifier"`
	// Block for logging configuration settings for the event bus.
	LogConfig EventBusLogConfigPtrOutput `pulumi:"logConfig"`
	// Name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure that the `name` matches the `eventSourceName`.
	//
	// The following arguments are optional:
	Name pulumi.StringOutput `pulumi:"name"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringOutput `pulumi:"region"`
	// Map of tags assigned to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapOutput `pulumi:"tags"`
	// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapOutput `pulumi:"tagsAll"`
}

// NewEventBus registers a new resource with the given unique name, arguments, and options.
func NewEventBus(ctx *pulumi.Context,
	name string, args *EventBusArgs, opts ...pulumi.ResourceOption) (*EventBus, error) {
	if args == nil {
		args = &EventBusArgs{}
	}

	opts = internal.PkgResourceDefaultOpts(opts)
	var resource EventBus
	err := ctx.RegisterResource("aws:cloudwatch/eventBus:EventBus", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetEventBus gets an existing EventBus resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetEventBus(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *EventBusState, opts ...pulumi.ResourceOption) (*EventBus, error) {
	var resource EventBus
	err := ctx.ReadResource("aws:cloudwatch/eventBus:EventBus", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering EventBus resources.
type eventBusState struct {
	// ARN of the event bus.
	Arn *string `pulumi:"arn"`
	// Configuration details of the Amazon SQS queue for EventBridge to use as a dead-letter queue (DLQ). This block supports the following arguments:
	DeadLetterConfig *EventBusDeadLetterConfig `pulumi:"deadLetterConfig"`
	// Event bus description.
	Description *string `pulumi:"description"`
	// Partner event source that the new event bus will be matched with. Must match `name`.
	EventSourceName *string `pulumi:"eventSourceName"`
	// Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
	KmsKeyIdentifier *string `pulumi:"kmsKeyIdentifier"`
	// Block for logging configuration settings for the event bus.
	LogConfig *EventBusLogConfig `pulumi:"logConfig"`
	// Name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure that the `name` matches the `eventSourceName`.
	//
	// The following arguments are optional:
	Name *string `pulumi:"name"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Map of tags assigned to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
	// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll map[string]string `pulumi:"tagsAll"`
}

type EventBusState struct {
	// ARN of the event bus.
	Arn pulumi.StringPtrInput
	// Configuration details of the Amazon SQS queue for EventBridge to use as a dead-letter queue (DLQ). This block supports the following arguments:
	DeadLetterConfig EventBusDeadLetterConfigPtrInput
	// Event bus description.
	Description pulumi.StringPtrInput
	// Partner event source that the new event bus will be matched with. Must match `name`.
	EventSourceName pulumi.StringPtrInput
	// Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
	KmsKeyIdentifier pulumi.StringPtrInput
	// Block for logging configuration settings for the event bus.
	LogConfig EventBusLogConfigPtrInput
	// Name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure that the `name` matches the `eventSourceName`.
	//
	// The following arguments are optional:
	Name pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Map of tags assigned to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
	// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
	TagsAll pulumi.StringMapInput
}

func (EventBusState) ElementType() reflect.Type {
	return reflect.TypeOf((*eventBusState)(nil)).Elem()
}

type eventBusArgs struct {
	// Configuration details of the Amazon SQS queue for EventBridge to use as a dead-letter queue (DLQ). This block supports the following arguments:
	DeadLetterConfig *EventBusDeadLetterConfig `pulumi:"deadLetterConfig"`
	// Event bus description.
	Description *string `pulumi:"description"`
	// Partner event source that the new event bus will be matched with. Must match `name`.
	EventSourceName *string `pulumi:"eventSourceName"`
	// Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
	KmsKeyIdentifier *string `pulumi:"kmsKeyIdentifier"`
	// Block for logging configuration settings for the event bus.
	LogConfig *EventBusLogConfig `pulumi:"logConfig"`
	// Name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure that the `name` matches the `eventSourceName`.
	//
	// The following arguments are optional:
	Name *string `pulumi:"name"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Map of tags assigned to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags map[string]string `pulumi:"tags"`
}

// The set of arguments for constructing a EventBus resource.
type EventBusArgs struct {
	// Configuration details of the Amazon SQS queue for EventBridge to use as a dead-letter queue (DLQ). This block supports the following arguments:
	DeadLetterConfig EventBusDeadLetterConfigPtrInput
	// Event bus description.
	Description pulumi.StringPtrInput
	// Partner event source that the new event bus will be matched with. Must match `name`.
	EventSourceName pulumi.StringPtrInput
	// Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
	KmsKeyIdentifier pulumi.StringPtrInput
	// Block for logging configuration settings for the event bus.
	LogConfig EventBusLogConfigPtrInput
	// Name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure that the `name` matches the `eventSourceName`.
	//
	// The following arguments are optional:
	Name pulumi.StringPtrInput
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput
	// Map of tags assigned to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
	Tags pulumi.StringMapInput
}

func (EventBusArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*eventBusArgs)(nil)).Elem()
}

type EventBusInput interface {
	pulumi.Input

	ToEventBusOutput() EventBusOutput
	ToEventBusOutputWithContext(ctx context.Context) EventBusOutput
}

func (*EventBus) ElementType() reflect.Type {
	return reflect.TypeOf((**EventBus)(nil)).Elem()
}

func (i *EventBus) ToEventBusOutput() EventBusOutput {
	return i.ToEventBusOutputWithContext(context.Background())
}

func (i *EventBus) ToEventBusOutputWithContext(ctx context.Context) EventBusOutput {
	return pulumi.ToOutputWithContext(ctx, i).(EventBusOutput)
}

// EventBusArrayInput is an input type that accepts EventBusArray and EventBusArrayOutput values.
// You can construct a concrete instance of `EventBusArrayInput` via:
//
//	EventBusArray{ EventBusArgs{...} }
type EventBusArrayInput interface {
	pulumi.Input

	ToEventBusArrayOutput() EventBusArrayOutput
	ToEventBusArrayOutputWithContext(context.Context) EventBusArrayOutput
}

type EventBusArray []EventBusInput

func (EventBusArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*EventBus)(nil)).Elem()
}

func (i EventBusArray) ToEventBusArrayOutput() EventBusArrayOutput {
	return i.ToEventBusArrayOutputWithContext(context.Background())
}

func (i EventBusArray) ToEventBusArrayOutputWithContext(ctx context.Context) EventBusArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(EventBusArrayOutput)
}

// EventBusMapInput is an input type that accepts EventBusMap and EventBusMapOutput values.
// You can construct a concrete instance of `EventBusMapInput` via:
//
//	EventBusMap{ "key": EventBusArgs{...} }
type EventBusMapInput interface {
	pulumi.Input

	ToEventBusMapOutput() EventBusMapOutput
	ToEventBusMapOutputWithContext(context.Context) EventBusMapOutput
}

type EventBusMap map[string]EventBusInput

func (EventBusMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*EventBus)(nil)).Elem()
}

func (i EventBusMap) ToEventBusMapOutput() EventBusMapOutput {
	return i.ToEventBusMapOutputWithContext(context.Background())
}

func (i EventBusMap) ToEventBusMapOutputWithContext(ctx context.Context) EventBusMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(EventBusMapOutput)
}

type EventBusOutput struct{ *pulumi.OutputState }

func (EventBusOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**EventBus)(nil)).Elem()
}

func (o EventBusOutput) ToEventBusOutput() EventBusOutput {
	return o
}

func (o EventBusOutput) ToEventBusOutputWithContext(ctx context.Context) EventBusOutput {
	return o
}

// ARN of the event bus.
func (o EventBusOutput) Arn() pulumi.StringOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringOutput { return v.Arn }).(pulumi.StringOutput)
}

// Configuration details of the Amazon SQS queue for EventBridge to use as a dead-letter queue (DLQ). This block supports the following arguments:
func (o EventBusOutput) DeadLetterConfig() EventBusDeadLetterConfigPtrOutput {
	return o.ApplyT(func(v *EventBus) EventBusDeadLetterConfigPtrOutput { return v.DeadLetterConfig }).(EventBusDeadLetterConfigPtrOutput)
}

// Event bus description.
func (o EventBusOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// Partner event source that the new event bus will be matched with. Must match `name`.
func (o EventBusOutput) EventSourceName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringPtrOutput { return v.EventSourceName }).(pulumi.StringPtrOutput)
}

// Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
func (o EventBusOutput) KmsKeyIdentifier() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringPtrOutput { return v.KmsKeyIdentifier }).(pulumi.StringPtrOutput)
}

// Block for logging configuration settings for the event bus.
func (o EventBusOutput) LogConfig() EventBusLogConfigPtrOutput {
	return o.ApplyT(func(v *EventBus) EventBusLogConfigPtrOutput { return v.LogConfig }).(EventBusLogConfigPtrOutput)
}

// Name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure that the `name` matches the `eventSourceName`.
//
// The following arguments are optional:
func (o EventBusOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
func (o EventBusOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringOutput { return v.Region }).(pulumi.StringOutput)
}

// Map of tags assigned to the resource. If configured with a provider `defaultTags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
func (o EventBusOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringMapOutput { return v.Tags }).(pulumi.StringMapOutput)
}

// Map of tags assigned to the resource, including those inherited from the provider `defaultTags` configuration block.
func (o EventBusOutput) TagsAll() pulumi.StringMapOutput {
	return o.ApplyT(func(v *EventBus) pulumi.StringMapOutput { return v.TagsAll }).(pulumi.StringMapOutput)
}

type EventBusArrayOutput struct{ *pulumi.OutputState }

func (EventBusArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*EventBus)(nil)).Elem()
}

func (o EventBusArrayOutput) ToEventBusArrayOutput() EventBusArrayOutput {
	return o
}

func (o EventBusArrayOutput) ToEventBusArrayOutputWithContext(ctx context.Context) EventBusArrayOutput {
	return o
}

func (o EventBusArrayOutput) Index(i pulumi.IntInput) EventBusOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *EventBus {
		return vs[0].([]*EventBus)[vs[1].(int)]
	}).(EventBusOutput)
}

type EventBusMapOutput struct{ *pulumi.OutputState }

func (EventBusMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*EventBus)(nil)).Elem()
}

func (o EventBusMapOutput) ToEventBusMapOutput() EventBusMapOutput {
	return o
}

func (o EventBusMapOutput) ToEventBusMapOutputWithContext(ctx context.Context) EventBusMapOutput {
	return o
}

func (o EventBusMapOutput) MapIndex(k pulumi.StringInput) EventBusOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *EventBus {
		return vs[0].(map[string]*EventBus)[vs[1].(string)]
	}).(EventBusOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*EventBusInput)(nil)).Elem(), &EventBus{})
	pulumi.RegisterInputType(reflect.TypeOf((*EventBusArrayInput)(nil)).Elem(), EventBusArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*EventBusMapInput)(nil)).Elem(), EventBusMap{})
	pulumi.RegisterOutputType(EventBusOutput{})
	pulumi.RegisterOutputType(EventBusArrayOutput{})
	pulumi.RegisterOutputType(EventBusMapOutput{})
}
