// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package aws

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetAccountId(ctx *pulumi.Context, opts ...pulumi.InvokeOption) (*GetAccountIdResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetAccountIdResult
	err := ctx.Invoke("aws-native:index:getAccountId", nil, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type GetAccountIdResult struct {
	AccountId string `pulumi:"accountId"`
}

func GetAccountIdOutput(ctx *pulumi.Context, opts ...pulumi.InvokeOption) GetAccountIdResultOutput {
	return pulumi.ToOutput(0).ApplyT(func(int) (GetAccountIdResultOutput, error) {
		options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
		return ctx.InvokeOutput("aws-native:index:getAccountId", nil, GetAccountIdResultOutput{}, options).(GetAccountIdResultOutput), nil
	}).(GetAccountIdResultOutput)
}

type GetAccountIdResultOutput struct{ *pulumi.OutputState }

func (GetAccountIdResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetAccountIdResult)(nil)).Elem()
}

func (o GetAccountIdResultOutput) ToGetAccountIdResultOutput() GetAccountIdResultOutput {
	return o
}

func (o GetAccountIdResultOutput) ToGetAccountIdResultOutputWithContext(ctx context.Context) GetAccountIdResultOutput {
	return o
}

func (o GetAccountIdResultOutput) AccountId() pulumi.StringOutput {
	return o.ApplyT(func(v GetAccountIdResult) string { return v.AccountId }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(GetAccountIdResultOutput{})
}
