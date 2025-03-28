// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package scheduler

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Definition of AWS::Scheduler::Schedule Resource Type
type Schedule struct {
	pulumi.CustomResourceState

	// The Amazon Resource Name (ARN) of the schedule.
	Arn pulumi.StringOutput `pulumi:"arn"`
	// The description of the schedule.
	Description pulumi.StringPtrOutput `pulumi:"description"`
	// The date, in UTC, before which the schedule can invoke its target. Depending on the schedule's recurrence expression, invocations might stop on, or before, the EndDate you specify.
	EndDate pulumi.StringPtrOutput `pulumi:"endDate"`
	// Allows you to configure a time window during which EventBridge Scheduler invokes the schedule.
	FlexibleTimeWindow ScheduleFlexibleTimeWindowOutput `pulumi:"flexibleTimeWindow"`
	// The name of the schedule group to associate with this schedule. If you omit this, the default schedule group is used.
	GroupName pulumi.StringPtrOutput `pulumi:"groupName"`
	// The ARN for a KMS Key that will be used to encrypt customer data.
	KmsKeyArn pulumi.StringPtrOutput `pulumi:"kmsKeyArn"`
	// The name of the schedule.
	Name pulumi.StringPtrOutput `pulumi:"name"`
	// The scheduling expression.
	ScheduleExpression pulumi.StringOutput `pulumi:"scheduleExpression"`
	// The timezone in which the scheduling expression is evaluated.
	ScheduleExpressionTimezone pulumi.StringPtrOutput `pulumi:"scheduleExpressionTimezone"`
	// The date, in UTC, after which the schedule can begin invoking its target. Depending on the schedule's recurrence expression, invocations might occur on, or after, the StartDate you specify.
	StartDate pulumi.StringPtrOutput `pulumi:"startDate"`
	// Specifies whether the schedule is enabled or disabled.
	//
	// *Allowed Values* : `ENABLED` | `DISABLED`
	State ScheduleStateEnumPtrOutput `pulumi:"state"`
	// The schedule's target details.
	Target ScheduleTargetOutput `pulumi:"target"`
}

// NewSchedule registers a new resource with the given unique name, arguments, and options.
func NewSchedule(ctx *pulumi.Context,
	name string, args *ScheduleArgs, opts ...pulumi.ResourceOption) (*Schedule, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.FlexibleTimeWindow == nil {
		return nil, errors.New("invalid value for required argument 'FlexibleTimeWindow'")
	}
	if args.ScheduleExpression == nil {
		return nil, errors.New("invalid value for required argument 'ScheduleExpression'")
	}
	if args.Target == nil {
		return nil, errors.New("invalid value for required argument 'Target'")
	}
	replaceOnChanges := pulumi.ReplaceOnChanges([]string{
		"name",
	})
	opts = append(opts, replaceOnChanges)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource Schedule
	err := ctx.RegisterResource("aws-native:scheduler:Schedule", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetSchedule gets an existing Schedule resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetSchedule(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *ScheduleState, opts ...pulumi.ResourceOption) (*Schedule, error) {
	var resource Schedule
	err := ctx.ReadResource("aws-native:scheduler:Schedule", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering Schedule resources.
type scheduleState struct {
}

type ScheduleState struct {
}

func (ScheduleState) ElementType() reflect.Type {
	return reflect.TypeOf((*scheduleState)(nil)).Elem()
}

type scheduleArgs struct {
	// The description of the schedule.
	Description *string `pulumi:"description"`
	// The date, in UTC, before which the schedule can invoke its target. Depending on the schedule's recurrence expression, invocations might stop on, or before, the EndDate you specify.
	EndDate *string `pulumi:"endDate"`
	// Allows you to configure a time window during which EventBridge Scheduler invokes the schedule.
	FlexibleTimeWindow ScheduleFlexibleTimeWindow `pulumi:"flexibleTimeWindow"`
	// The name of the schedule group to associate with this schedule. If you omit this, the default schedule group is used.
	GroupName *string `pulumi:"groupName"`
	// The ARN for a KMS Key that will be used to encrypt customer data.
	KmsKeyArn *string `pulumi:"kmsKeyArn"`
	// The name of the schedule.
	Name *string `pulumi:"name"`
	// The scheduling expression.
	ScheduleExpression string `pulumi:"scheduleExpression"`
	// The timezone in which the scheduling expression is evaluated.
	ScheduleExpressionTimezone *string `pulumi:"scheduleExpressionTimezone"`
	// The date, in UTC, after which the schedule can begin invoking its target. Depending on the schedule's recurrence expression, invocations might occur on, or after, the StartDate you specify.
	StartDate *string `pulumi:"startDate"`
	// Specifies whether the schedule is enabled or disabled.
	//
	// *Allowed Values* : `ENABLED` | `DISABLED`
	State *ScheduleStateEnum `pulumi:"state"`
	// The schedule's target details.
	Target ScheduleTarget `pulumi:"target"`
}

// The set of arguments for constructing a Schedule resource.
type ScheduleArgs struct {
	// The description of the schedule.
	Description pulumi.StringPtrInput
	// The date, in UTC, before which the schedule can invoke its target. Depending on the schedule's recurrence expression, invocations might stop on, or before, the EndDate you specify.
	EndDate pulumi.StringPtrInput
	// Allows you to configure a time window during which EventBridge Scheduler invokes the schedule.
	FlexibleTimeWindow ScheduleFlexibleTimeWindowInput
	// The name of the schedule group to associate with this schedule. If you omit this, the default schedule group is used.
	GroupName pulumi.StringPtrInput
	// The ARN for a KMS Key that will be used to encrypt customer data.
	KmsKeyArn pulumi.StringPtrInput
	// The name of the schedule.
	Name pulumi.StringPtrInput
	// The scheduling expression.
	ScheduleExpression pulumi.StringInput
	// The timezone in which the scheduling expression is evaluated.
	ScheduleExpressionTimezone pulumi.StringPtrInput
	// The date, in UTC, after which the schedule can begin invoking its target. Depending on the schedule's recurrence expression, invocations might occur on, or after, the StartDate you specify.
	StartDate pulumi.StringPtrInput
	// Specifies whether the schedule is enabled or disabled.
	//
	// *Allowed Values* : `ENABLED` | `DISABLED`
	State ScheduleStateEnumPtrInput
	// The schedule's target details.
	Target ScheduleTargetInput
}

func (ScheduleArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*scheduleArgs)(nil)).Elem()
}

type ScheduleInput interface {
	pulumi.Input

	ToScheduleOutput() ScheduleOutput
	ToScheduleOutputWithContext(ctx context.Context) ScheduleOutput
}

func (*Schedule) ElementType() reflect.Type {
	return reflect.TypeOf((**Schedule)(nil)).Elem()
}

func (i *Schedule) ToScheduleOutput() ScheduleOutput {
	return i.ToScheduleOutputWithContext(context.Background())
}

func (i *Schedule) ToScheduleOutputWithContext(ctx context.Context) ScheduleOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ScheduleOutput)
}

type ScheduleOutput struct{ *pulumi.OutputState }

func (ScheduleOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**Schedule)(nil)).Elem()
}

func (o ScheduleOutput) ToScheduleOutput() ScheduleOutput {
	return o
}

func (o ScheduleOutput) ToScheduleOutputWithContext(ctx context.Context) ScheduleOutput {
	return o
}

// The Amazon Resource Name (ARN) of the schedule.
func (o ScheduleOutput) Arn() pulumi.StringOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringOutput { return v.Arn }).(pulumi.StringOutput)
}

// The description of the schedule.
func (o ScheduleOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.Description }).(pulumi.StringPtrOutput)
}

// The date, in UTC, before which the schedule can invoke its target. Depending on the schedule's recurrence expression, invocations might stop on, or before, the EndDate you specify.
func (o ScheduleOutput) EndDate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.EndDate }).(pulumi.StringPtrOutput)
}

// Allows you to configure a time window during which EventBridge Scheduler invokes the schedule.
func (o ScheduleOutput) FlexibleTimeWindow() ScheduleFlexibleTimeWindowOutput {
	return o.ApplyT(func(v *Schedule) ScheduleFlexibleTimeWindowOutput { return v.FlexibleTimeWindow }).(ScheduleFlexibleTimeWindowOutput)
}

// The name of the schedule group to associate with this schedule. If you omit this, the default schedule group is used.
func (o ScheduleOutput) GroupName() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.GroupName }).(pulumi.StringPtrOutput)
}

// The ARN for a KMS Key that will be used to encrypt customer data.
func (o ScheduleOutput) KmsKeyArn() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.KmsKeyArn }).(pulumi.StringPtrOutput)
}

// The name of the schedule.
func (o ScheduleOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.Name }).(pulumi.StringPtrOutput)
}

// The scheduling expression.
func (o ScheduleOutput) ScheduleExpression() pulumi.StringOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringOutput { return v.ScheduleExpression }).(pulumi.StringOutput)
}

// The timezone in which the scheduling expression is evaluated.
func (o ScheduleOutput) ScheduleExpressionTimezone() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.ScheduleExpressionTimezone }).(pulumi.StringPtrOutput)
}

// The date, in UTC, after which the schedule can begin invoking its target. Depending on the schedule's recurrence expression, invocations might occur on, or after, the StartDate you specify.
func (o ScheduleOutput) StartDate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Schedule) pulumi.StringPtrOutput { return v.StartDate }).(pulumi.StringPtrOutput)
}

// Specifies whether the schedule is enabled or disabled.
//
// *Allowed Values* : `ENABLED` | `DISABLED`
func (o ScheduleOutput) State() ScheduleStateEnumPtrOutput {
	return o.ApplyT(func(v *Schedule) ScheduleStateEnumPtrOutput { return v.State }).(ScheduleStateEnumPtrOutput)
}

// The schedule's target details.
func (o ScheduleOutput) Target() ScheduleTargetOutput {
	return o.ApplyT(func(v *Schedule) ScheduleTargetOutput { return v.Target }).(ScheduleTargetOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*ScheduleInput)(nil)).Elem(), &Schedule{})
	pulumi.RegisterOutputType(ScheduleOutput{})
}
