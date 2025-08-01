// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package aws

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Use this data source to create a Service Principal Name for a service in a given region. Service Principal Names should always end in the standard global format: `{servicename}.amazonaws.com`. However, in some AWS partitions, AWS may expect a different format.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := aws.GetServicePrincipal(ctx, &aws.GetServicePrincipalArgs{
//				ServiceName: "s3",
//			}, nil)
//			if err != nil {
//				return err
//			}
//			_, err = aws.GetServicePrincipal(ctx, &aws.GetServicePrincipalArgs{
//				ServiceName: "s3",
//				Region:      pulumi.StringRef("us-iso-east-1"),
//			}, nil)
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
func GetServicePrincipal(ctx *pulumi.Context, args *GetServicePrincipalArgs, opts ...pulumi.InvokeOption) (*GetServicePrincipalResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetServicePrincipalResult
	err := ctx.Invoke("aws:index/getServicePrincipal:getServicePrincipal", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getServicePrincipal.
type GetServicePrincipalArgs struct {
	// Region you'd like the SPN for. Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Name of the service you want to generate a Service Principal Name for.
	ServiceName string `pulumi:"serviceName"`
}

// A collection of values returned by getServicePrincipal.
type GetServicePrincipalResult struct {
	// Identifier of the current Service Principal (compound of service, Region and suffix). (e.g. `logs.us-east-1.amazonaws.com`in AWS Commercial, `logs.cn-north-1.amazonaws.com.cn` in AWS China).
	Id string `pulumi:"id"`
	// Service Principal Name (e.g., `logs.amazonaws.com` in AWS Commercial, `logs.amazonaws.com.cn` in AWS China).
	Name        string `pulumi:"name"`
	Region      string `pulumi:"region"`
	ServiceName string `pulumi:"serviceName"`
	// Suffix of the SPN (e.g., `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China).
	Suffix string `pulumi:"suffix"`
}

func GetServicePrincipalOutput(ctx *pulumi.Context, args GetServicePrincipalOutputArgs, opts ...pulumi.InvokeOption) GetServicePrincipalResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (GetServicePrincipalResultOutput, error) {
			args := v.(GetServicePrincipalArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:index/getServicePrincipal:getServicePrincipal", args, GetServicePrincipalResultOutput{}, options).(GetServicePrincipalResultOutput), nil
		}).(GetServicePrincipalResultOutput)
}

// A collection of arguments for invoking getServicePrincipal.
type GetServicePrincipalOutputArgs struct {
	// Region you'd like the SPN for. Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput `pulumi:"region"`
	// Name of the service you want to generate a Service Principal Name for.
	ServiceName pulumi.StringInput `pulumi:"serviceName"`
}

func (GetServicePrincipalOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetServicePrincipalArgs)(nil)).Elem()
}

// A collection of values returned by getServicePrincipal.
type GetServicePrincipalResultOutput struct{ *pulumi.OutputState }

func (GetServicePrincipalResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetServicePrincipalResult)(nil)).Elem()
}

func (o GetServicePrincipalResultOutput) ToGetServicePrincipalResultOutput() GetServicePrincipalResultOutput {
	return o
}

func (o GetServicePrincipalResultOutput) ToGetServicePrincipalResultOutputWithContext(ctx context.Context) GetServicePrincipalResultOutput {
	return o
}

// Identifier of the current Service Principal (compound of service, Region and suffix). (e.g. `logs.us-east-1.amazonaws.com`in AWS Commercial, `logs.cn-north-1.amazonaws.com.cn` in AWS China).
func (o GetServicePrincipalResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetServicePrincipalResult) string { return v.Id }).(pulumi.StringOutput)
}

// Service Principal Name (e.g., `logs.amazonaws.com` in AWS Commercial, `logs.amazonaws.com.cn` in AWS China).
func (o GetServicePrincipalResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v GetServicePrincipalResult) string { return v.Name }).(pulumi.StringOutput)
}

func (o GetServicePrincipalResultOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v GetServicePrincipalResult) string { return v.Region }).(pulumi.StringOutput)
}

func (o GetServicePrincipalResultOutput) ServiceName() pulumi.StringOutput {
	return o.ApplyT(func(v GetServicePrincipalResult) string { return v.ServiceName }).(pulumi.StringOutput)
}

// Suffix of the SPN (e.g., `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China).
func (o GetServicePrincipalResultOutput) Suffix() pulumi.StringOutput {
	return o.ApplyT(func(v GetServicePrincipalResult) string { return v.Suffix }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(GetServicePrincipalResultOutput{})
}
