// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package iam

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// This data source can be used to fetch information about IAM access keys of a
// specific IAM user.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := iam.GetAccessKeys(ctx, &iam.GetAccessKeysArgs{
//				User: "an_example_user_name",
//			}, nil)
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
func GetAccessKeys(ctx *pulumi.Context, args *GetAccessKeysArgs, opts ...pulumi.InvokeOption) (*GetAccessKeysResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetAccessKeysResult
	err := ctx.Invoke("aws:iam/getAccessKeys:getAccessKeys", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getAccessKeys.
type GetAccessKeysArgs struct {
	// Name of the IAM user associated with the access keys.
	User string `pulumi:"user"`
}

// A collection of values returned by getAccessKeys.
type GetAccessKeysResult struct {
	// List of the IAM access keys associated with the specified user. See below.
	AccessKeys []GetAccessKeysAccessKey `pulumi:"accessKeys"`
	// The provider-assigned unique ID for this managed resource.
	Id   string `pulumi:"id"`
	User string `pulumi:"user"`
}

func GetAccessKeysOutput(ctx *pulumi.Context, args GetAccessKeysOutputArgs, opts ...pulumi.InvokeOption) GetAccessKeysResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (GetAccessKeysResultOutput, error) {
			args := v.(GetAccessKeysArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:iam/getAccessKeys:getAccessKeys", args, GetAccessKeysResultOutput{}, options).(GetAccessKeysResultOutput), nil
		}).(GetAccessKeysResultOutput)
}

// A collection of arguments for invoking getAccessKeys.
type GetAccessKeysOutputArgs struct {
	// Name of the IAM user associated with the access keys.
	User pulumi.StringInput `pulumi:"user"`
}

func (GetAccessKeysOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetAccessKeysArgs)(nil)).Elem()
}

// A collection of values returned by getAccessKeys.
type GetAccessKeysResultOutput struct{ *pulumi.OutputState }

func (GetAccessKeysResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetAccessKeysResult)(nil)).Elem()
}

func (o GetAccessKeysResultOutput) ToGetAccessKeysResultOutput() GetAccessKeysResultOutput {
	return o
}

func (o GetAccessKeysResultOutput) ToGetAccessKeysResultOutputWithContext(ctx context.Context) GetAccessKeysResultOutput {
	return o
}

// List of the IAM access keys associated with the specified user. See below.
func (o GetAccessKeysResultOutput) AccessKeys() GetAccessKeysAccessKeyArrayOutput {
	return o.ApplyT(func(v GetAccessKeysResult) []GetAccessKeysAccessKey { return v.AccessKeys }).(GetAccessKeysAccessKeyArrayOutput)
}

// The provider-assigned unique ID for this managed resource.
func (o GetAccessKeysResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetAccessKeysResult) string { return v.Id }).(pulumi.StringOutput)
}

func (o GetAccessKeysResultOutput) User() pulumi.StringOutput {
	return o.ApplyT(func(v GetAccessKeysResult) string { return v.User }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(GetAccessKeysResultOutput{})
}
