// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package network

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Lists Active Security User Rules in a network manager.
//
// Uses Azure REST API version 2022-04-01-preview.
//
// Other available API versions: 2021-05-01-preview.
func ListActiveSecurityUserRules(ctx *pulumi.Context, args *ListActiveSecurityUserRulesArgs, opts ...pulumi.InvokeOption) (*ListActiveSecurityUserRulesResult, error) {
	opts = utilities.PkgInvokeDefaultOpts(opts)
	var rv ListActiveSecurityUserRulesResult
	err := ctx.Invoke("azure-native:network:listActiveSecurityUserRules", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type ListActiveSecurityUserRulesArgs struct {
	// The name of the network manager.
	NetworkManagerName string `pulumi:"networkManagerName"`
	// List of regions.
	Regions []string `pulumi:"regions"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// When present, the value can be passed to a subsequent query call (together with the same query and scopes used in the current request) to retrieve the next page of data.
	SkipToken *string `pulumi:"skipToken"`
}

// Result of the request to list active security user rules. It contains a list of active security user rules and a skiptoken to get the next set of results.
type ListActiveSecurityUserRulesResult struct {
	// When present, the value can be passed to a subsequent query call (together with the same query and scopes used in the current request) to retrieve the next page of data.
	SkipToken *string `pulumi:"skipToken"`
	// Gets a page of active security user rules.
	Value []interface{} `pulumi:"value"`
}

func ListActiveSecurityUserRulesOutput(ctx *pulumi.Context, args ListActiveSecurityUserRulesOutputArgs, opts ...pulumi.InvokeOption) ListActiveSecurityUserRulesResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (ListActiveSecurityUserRulesResultOutput, error) {
			args := v.(ListActiveSecurityUserRulesArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: utilities.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("azure-native:network:listActiveSecurityUserRules", args, ListActiveSecurityUserRulesResultOutput{}, options).(ListActiveSecurityUserRulesResultOutput), nil
		}).(ListActiveSecurityUserRulesResultOutput)
}

type ListActiveSecurityUserRulesOutputArgs struct {
	// The name of the network manager.
	NetworkManagerName pulumi.StringInput `pulumi:"networkManagerName"`
	// List of regions.
	Regions pulumi.StringArrayInput `pulumi:"regions"`
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput `pulumi:"resourceGroupName"`
	// When present, the value can be passed to a subsequent query call (together with the same query and scopes used in the current request) to retrieve the next page of data.
	SkipToken pulumi.StringPtrInput `pulumi:"skipToken"`
}

func (ListActiveSecurityUserRulesOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*ListActiveSecurityUserRulesArgs)(nil)).Elem()
}

// Result of the request to list active security user rules. It contains a list of active security user rules and a skiptoken to get the next set of results.
type ListActiveSecurityUserRulesResultOutput struct{ *pulumi.OutputState }

func (ListActiveSecurityUserRulesResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*ListActiveSecurityUserRulesResult)(nil)).Elem()
}

func (o ListActiveSecurityUserRulesResultOutput) ToListActiveSecurityUserRulesResultOutput() ListActiveSecurityUserRulesResultOutput {
	return o
}

func (o ListActiveSecurityUserRulesResultOutput) ToListActiveSecurityUserRulesResultOutputWithContext(ctx context.Context) ListActiveSecurityUserRulesResultOutput {
	return o
}

// When present, the value can be passed to a subsequent query call (together with the same query and scopes used in the current request) to retrieve the next page of data.
func (o ListActiveSecurityUserRulesResultOutput) SkipToken() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ListActiveSecurityUserRulesResult) *string { return v.SkipToken }).(pulumi.StringPtrOutput)
}

// Gets a page of active security user rules.
func (o ListActiveSecurityUserRulesResultOutput) Value() pulumi.ArrayOutput {
	return o.ApplyT(func(v ListActiveSecurityUserRulesResult) []interface{} { return v.Value }).(pulumi.ArrayOutput)
}

func init() {
	pulumi.RegisterOutputType(ListActiveSecurityUserRulesResultOutput{})
}
