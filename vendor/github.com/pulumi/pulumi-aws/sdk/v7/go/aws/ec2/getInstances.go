// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package ec2

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Use this data source to get IDs or IPs of Amazon EC2 instances to be referenced elsewhere,
// e.g., to allow easier migration from another management solution
// or to make it easier for an operator to connect through bastion host(s).
//
// > **Note:** It's strongly discouraged to use this data source for querying ephemeral
// instances (e.g., managed via autoscaling group), as the output may change at any time
// and you'd need to re-run `apply` every time an instance comes up or dies.
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
//			test, err := ec2.GetInstances(ctx, &ec2.GetInstancesArgs{
//				InstanceTags: map[string]interface{}{
//					"Role": "HardWorker",
//				},
//				Filters: []ec2.GetInstancesFilter{
//					{
//						Name: "instance.group-id",
//						Values: []string{
//							"sg-12345678",
//						},
//					},
//				},
//				InstanceStateNames: []string{
//					"running",
//					"stopped",
//				},
//			}, nil)
//			if err != nil {
//				return err
//			}
//			var testEip []*ec2.Eip
//			for index := 0; index < int(len(test.Ids)); index++ {
//				key0 := index
//				val0 := index
//				__res, err := ec2.NewEip(ctx, fmt.Sprintf("test-%v", key0), &ec2.EipArgs{
//					Instance: pulumi.String(test.Ids[val0]),
//				})
//				if err != nil {
//					return err
//				}
//				testEip = append(testEip, __res)
//			}
//			return nil
//		})
//	}
//
// ```
func GetInstances(ctx *pulumi.Context, args *GetInstancesArgs, opts ...pulumi.InvokeOption) (*GetInstancesResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetInstancesResult
	err := ctx.Invoke("aws:ec2/getInstances:getInstances", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getInstances.
type GetInstancesArgs struct {
	// One or more name/value pairs to use as filters. There are
	// several valid keys, for a full reference, check out
	// [describe-instances in the AWS CLI reference][1].
	Filters []GetInstancesFilter `pulumi:"filters"`
	// List of instance states that should be applicable to the desired instances. The permitted values are: `pending, running, shutting-down, stopped, stopping, terminated`. The default value is `running`.
	InstanceStateNames []string `pulumi:"instanceStateNames"`
	// Map of tags, each pair of which must
	// exactly match a pair on desired instances.
	InstanceTags map[string]string `pulumi:"instanceTags"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

// A collection of values returned by getInstances.
type GetInstancesResult struct {
	Filters []GetInstancesFilter `pulumi:"filters"`
	// The provider-assigned unique ID for this managed resource.
	Id string `pulumi:"id"`
	// IDs of instances found through the filter
	Ids                []string          `pulumi:"ids"`
	InstanceStateNames []string          `pulumi:"instanceStateNames"`
	InstanceTags       map[string]string `pulumi:"instanceTags"`
	// IPv6 addresses of instances found through the filter
	Ipv6Addresses []string `pulumi:"ipv6Addresses"`
	// Private IP addresses of instances found through the filter
	PrivateIps []string `pulumi:"privateIps"`
	// Public IP addresses of instances found through the filter
	PublicIps []string `pulumi:"publicIps"`
	Region    string   `pulumi:"region"`
}

func GetInstancesOutput(ctx *pulumi.Context, args GetInstancesOutputArgs, opts ...pulumi.InvokeOption) GetInstancesResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (GetInstancesResultOutput, error) {
			args := v.(GetInstancesArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:ec2/getInstances:getInstances", args, GetInstancesResultOutput{}, options).(GetInstancesResultOutput), nil
		}).(GetInstancesResultOutput)
}

// A collection of arguments for invoking getInstances.
type GetInstancesOutputArgs struct {
	// One or more name/value pairs to use as filters. There are
	// several valid keys, for a full reference, check out
	// [describe-instances in the AWS CLI reference][1].
	Filters GetInstancesFilterArrayInput `pulumi:"filters"`
	// List of instance states that should be applicable to the desired instances. The permitted values are: `pending, running, shutting-down, stopped, stopping, terminated`. The default value is `running`.
	InstanceStateNames pulumi.StringArrayInput `pulumi:"instanceStateNames"`
	// Map of tags, each pair of which must
	// exactly match a pair on desired instances.
	InstanceTags pulumi.StringMapInput `pulumi:"instanceTags"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput `pulumi:"region"`
}

func (GetInstancesOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetInstancesArgs)(nil)).Elem()
}

// A collection of values returned by getInstances.
type GetInstancesResultOutput struct{ *pulumi.OutputState }

func (GetInstancesResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetInstancesResult)(nil)).Elem()
}

func (o GetInstancesResultOutput) ToGetInstancesResultOutput() GetInstancesResultOutput {
	return o
}

func (o GetInstancesResultOutput) ToGetInstancesResultOutputWithContext(ctx context.Context) GetInstancesResultOutput {
	return o
}

func (o GetInstancesResultOutput) Filters() GetInstancesFilterArrayOutput {
	return o.ApplyT(func(v GetInstancesResult) []GetInstancesFilter { return v.Filters }).(GetInstancesFilterArrayOutput)
}

// The provider-assigned unique ID for this managed resource.
func (o GetInstancesResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetInstancesResult) string { return v.Id }).(pulumi.StringOutput)
}

// IDs of instances found through the filter
func (o GetInstancesResultOutput) Ids() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetInstancesResult) []string { return v.Ids }).(pulumi.StringArrayOutput)
}

func (o GetInstancesResultOutput) InstanceStateNames() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetInstancesResult) []string { return v.InstanceStateNames }).(pulumi.StringArrayOutput)
}

func (o GetInstancesResultOutput) InstanceTags() pulumi.StringMapOutput {
	return o.ApplyT(func(v GetInstancesResult) map[string]string { return v.InstanceTags }).(pulumi.StringMapOutput)
}

// IPv6 addresses of instances found through the filter
func (o GetInstancesResultOutput) Ipv6Addresses() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetInstancesResult) []string { return v.Ipv6Addresses }).(pulumi.StringArrayOutput)
}

// Private IP addresses of instances found through the filter
func (o GetInstancesResultOutput) PrivateIps() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetInstancesResult) []string { return v.PrivateIps }).(pulumi.StringArrayOutput)
}

// Public IP addresses of instances found through the filter
func (o GetInstancesResultOutput) PublicIps() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetInstancesResult) []string { return v.PublicIps }).(pulumi.StringArrayOutput)
}

func (o GetInstancesResultOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v GetInstancesResult) string { return v.Region }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(GetInstancesResultOutput{})
}
