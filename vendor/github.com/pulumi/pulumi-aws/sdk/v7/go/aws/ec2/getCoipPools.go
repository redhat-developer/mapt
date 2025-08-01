// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package ec2

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provides information for multiple EC2 Customer-Owned IP Pools, such as their identifiers.
func GetCoipPools(ctx *pulumi.Context, args *GetCoipPoolsArgs, opts ...pulumi.InvokeOption) (*GetCoipPoolsResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetCoipPoolsResult
	err := ctx.Invoke("aws:ec2/getCoipPools:getCoipPools", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getCoipPools.
type GetCoipPoolsArgs struct {
	// Custom filter block as described below.
	//
	// More complex filters can be expressed using one or more `filter` sub-blocks,
	// which take the following arguments:
	Filters []GetCoipPoolsFilter `pulumi:"filters"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Mapping of tags, each pair of which must exactly match
	// a pair on the desired aws_ec2_coip_pools.
	Tags map[string]string `pulumi:"tags"`
}

// A collection of values returned by getCoipPools.
type GetCoipPoolsResult struct {
	Filters []GetCoipPoolsFilter `pulumi:"filters"`
	// The provider-assigned unique ID for this managed resource.
	Id string `pulumi:"id"`
	// Set of COIP Pool Identifiers
	PoolIds []string          `pulumi:"poolIds"`
	Region  string            `pulumi:"region"`
	Tags    map[string]string `pulumi:"tags"`
}

func GetCoipPoolsOutput(ctx *pulumi.Context, args GetCoipPoolsOutputArgs, opts ...pulumi.InvokeOption) GetCoipPoolsResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (GetCoipPoolsResultOutput, error) {
			args := v.(GetCoipPoolsArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:ec2/getCoipPools:getCoipPools", args, GetCoipPoolsResultOutput{}, options).(GetCoipPoolsResultOutput), nil
		}).(GetCoipPoolsResultOutput)
}

// A collection of arguments for invoking getCoipPools.
type GetCoipPoolsOutputArgs struct {
	// Custom filter block as described below.
	//
	// More complex filters can be expressed using one or more `filter` sub-blocks,
	// which take the following arguments:
	Filters GetCoipPoolsFilterArrayInput `pulumi:"filters"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput `pulumi:"region"`
	// Mapping of tags, each pair of which must exactly match
	// a pair on the desired aws_ec2_coip_pools.
	Tags pulumi.StringMapInput `pulumi:"tags"`
}

func (GetCoipPoolsOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetCoipPoolsArgs)(nil)).Elem()
}

// A collection of values returned by getCoipPools.
type GetCoipPoolsResultOutput struct{ *pulumi.OutputState }

func (GetCoipPoolsResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetCoipPoolsResult)(nil)).Elem()
}

func (o GetCoipPoolsResultOutput) ToGetCoipPoolsResultOutput() GetCoipPoolsResultOutput {
	return o
}

func (o GetCoipPoolsResultOutput) ToGetCoipPoolsResultOutputWithContext(ctx context.Context) GetCoipPoolsResultOutput {
	return o
}

func (o GetCoipPoolsResultOutput) Filters() GetCoipPoolsFilterArrayOutput {
	return o.ApplyT(func(v GetCoipPoolsResult) []GetCoipPoolsFilter { return v.Filters }).(GetCoipPoolsFilterArrayOutput)
}

// The provider-assigned unique ID for this managed resource.
func (o GetCoipPoolsResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetCoipPoolsResult) string { return v.Id }).(pulumi.StringOutput)
}

// Set of COIP Pool Identifiers
func (o GetCoipPoolsResultOutput) PoolIds() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetCoipPoolsResult) []string { return v.PoolIds }).(pulumi.StringArrayOutput)
}

func (o GetCoipPoolsResultOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v GetCoipPoolsResult) string { return v.Region }).(pulumi.StringOutput)
}

func (o GetCoipPoolsResultOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v GetCoipPoolsResult) map[string]string { return v.Tags }).(pulumi.StringMapOutput)
}

func init() {
	pulumi.RegisterOutputType(GetCoipPoolsResultOutput{})
}
