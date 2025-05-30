// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package scheduler

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Definition of AWS::Scheduler::ScheduleGroup Resource Type
func LookupScheduleGroup(ctx *pulumi.Context, args *LookupScheduleGroupArgs, opts ...pulumi.InvokeOption) (*LookupScheduleGroupResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv LookupScheduleGroupResult
	err := ctx.Invoke("aws-native:scheduler:getScheduleGroup", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type LookupScheduleGroupArgs struct {
	// The name of the schedule group.
	Name string `pulumi:"name"`
}

type LookupScheduleGroupResult struct {
	// The Amazon Resource Name (ARN) of the schedule group.
	Arn *string `pulumi:"arn"`
	// The time at which the schedule group was created.
	CreationDate *string `pulumi:"creationDate"`
	// The time at which the schedule group was last modified.
	LastModificationDate *string `pulumi:"lastModificationDate"`
	// Specifies the state of the schedule group.
	//
	// *Allowed Values* : `ACTIVE` | `DELETING`
	State *ScheduleGroupStateEnum `pulumi:"state"`
	// The list of tags to associate with the schedule group.
	Tags []aws.Tag `pulumi:"tags"`
}

func LookupScheduleGroupOutput(ctx *pulumi.Context, args LookupScheduleGroupOutputArgs, opts ...pulumi.InvokeOption) LookupScheduleGroupResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupScheduleGroupResultOutput, error) {
			args := v.(LookupScheduleGroupArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws-native:scheduler:getScheduleGroup", args, LookupScheduleGroupResultOutput{}, options).(LookupScheduleGroupResultOutput), nil
		}).(LookupScheduleGroupResultOutput)
}

type LookupScheduleGroupOutputArgs struct {
	// The name of the schedule group.
	Name pulumi.StringInput `pulumi:"name"`
}

func (LookupScheduleGroupOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupScheduleGroupArgs)(nil)).Elem()
}

type LookupScheduleGroupResultOutput struct{ *pulumi.OutputState }

func (LookupScheduleGroupResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupScheduleGroupResult)(nil)).Elem()
}

func (o LookupScheduleGroupResultOutput) ToLookupScheduleGroupResultOutput() LookupScheduleGroupResultOutput {
	return o
}

func (o LookupScheduleGroupResultOutput) ToLookupScheduleGroupResultOutputWithContext(ctx context.Context) LookupScheduleGroupResultOutput {
	return o
}

// The Amazon Resource Name (ARN) of the schedule group.
func (o LookupScheduleGroupResultOutput) Arn() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupScheduleGroupResult) *string { return v.Arn }).(pulumi.StringPtrOutput)
}

// The time at which the schedule group was created.
func (o LookupScheduleGroupResultOutput) CreationDate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupScheduleGroupResult) *string { return v.CreationDate }).(pulumi.StringPtrOutput)
}

// The time at which the schedule group was last modified.
func (o LookupScheduleGroupResultOutput) LastModificationDate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupScheduleGroupResult) *string { return v.LastModificationDate }).(pulumi.StringPtrOutput)
}

// Specifies the state of the schedule group.
//
// *Allowed Values* : `ACTIVE` | `DELETING`
func (o LookupScheduleGroupResultOutput) State() ScheduleGroupStateEnumPtrOutput {
	return o.ApplyT(func(v LookupScheduleGroupResult) *ScheduleGroupStateEnum { return v.State }).(ScheduleGroupStateEnumPtrOutput)
}

// The list of tags to associate with the schedule group.
func (o LookupScheduleGroupResultOutput) Tags() aws.TagArrayOutput {
	return o.ApplyT(func(v LookupScheduleGroupResult) []aws.Tag { return v.Tags }).(aws.TagArrayOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupScheduleGroupResultOutput{})
}
