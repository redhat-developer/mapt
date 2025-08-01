// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package eks

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Retrieve EKS Clusters list
func GetClusters(ctx *pulumi.Context, args *GetClustersArgs, opts ...pulumi.InvokeOption) (*GetClustersResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetClustersResult
	err := ctx.Invoke("aws:eks/getClusters:getClusters", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getClusters.
type GetClustersArgs struct {
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
}

// A collection of values returned by getClusters.
type GetClustersResult struct {
	// The provider-assigned unique ID for this managed resource.
	Id string `pulumi:"id"`
	// Set of EKS clusters names
	Names  []string `pulumi:"names"`
	Region string   `pulumi:"region"`
}

func GetClustersOutput(ctx *pulumi.Context, args GetClustersOutputArgs, opts ...pulumi.InvokeOption) GetClustersResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (GetClustersResultOutput, error) {
			args := v.(GetClustersArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:eks/getClusters:getClusters", args, GetClustersResultOutput{}, options).(GetClustersResultOutput), nil
		}).(GetClustersResultOutput)
}

// A collection of arguments for invoking getClusters.
type GetClustersOutputArgs struct {
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput `pulumi:"region"`
}

func (GetClustersOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetClustersArgs)(nil)).Elem()
}

// A collection of values returned by getClusters.
type GetClustersResultOutput struct{ *pulumi.OutputState }

func (GetClustersResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetClustersResult)(nil)).Elem()
}

func (o GetClustersResultOutput) ToGetClustersResultOutput() GetClustersResultOutput {
	return o
}

func (o GetClustersResultOutput) ToGetClustersResultOutputWithContext(ctx context.Context) GetClustersResultOutput {
	return o
}

// The provider-assigned unique ID for this managed resource.
func (o GetClustersResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetClustersResult) string { return v.Id }).(pulumi.StringOutput)
}

// Set of EKS clusters names
func (o GetClustersResultOutput) Names() pulumi.StringArrayOutput {
	return o.ApplyT(func(v GetClustersResult) []string { return v.Names }).(pulumi.StringArrayOutput)
}

func (o GetClustersResultOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v GetClustersResult) string { return v.Region }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(GetClustersResultOutput{})
}
