// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package network

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Gets the NSP logging configuration.
//
// Uses Azure REST API version 2024-06-01-preview.
func LookupNetworkSecurityPerimeterLoggingConfiguration(ctx *pulumi.Context, args *LookupNetworkSecurityPerimeterLoggingConfigurationArgs, opts ...pulumi.InvokeOption) (*LookupNetworkSecurityPerimeterLoggingConfigurationResult, error) {
	opts = utilities.PkgInvokeDefaultOpts(opts)
	var rv LookupNetworkSecurityPerimeterLoggingConfigurationResult
	err := ctx.Invoke("azure-native:network:getNetworkSecurityPerimeterLoggingConfiguration", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type LookupNetworkSecurityPerimeterLoggingConfigurationArgs struct {
	// The name of the NSP logging configuration. Accepts 'instance' as name.
	LoggingConfigurationName string `pulumi:"loggingConfigurationName"`
	// The name of the network security perimeter.
	NetworkSecurityPerimeterName string `pulumi:"networkSecurityPerimeterName"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
}

// The NSP logging configuration
type LookupNetworkSecurityPerimeterLoggingConfigurationResult struct {
	// A unique read-only string that changes whenever the resource is updated.
	Etag string `pulumi:"etag"`
	// Resource ID.
	Id string `pulumi:"id"`
	// Resource name.
	Name string `pulumi:"name"`
	// Properties of the NSP logging configuration.
	Properties NspLoggingConfigurationPropertiesResponse `pulumi:"properties"`
	// Resource type.
	Type string `pulumi:"type"`
}

func LookupNetworkSecurityPerimeterLoggingConfigurationOutput(ctx *pulumi.Context, args LookupNetworkSecurityPerimeterLoggingConfigurationOutputArgs, opts ...pulumi.InvokeOption) LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput, error) {
			args := v.(LookupNetworkSecurityPerimeterLoggingConfigurationArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: utilities.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("azure-native:network:getNetworkSecurityPerimeterLoggingConfiguration", args, LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput{}, options).(LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput), nil
		}).(LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput)
}

type LookupNetworkSecurityPerimeterLoggingConfigurationOutputArgs struct {
	// The name of the NSP logging configuration. Accepts 'instance' as name.
	LoggingConfigurationName pulumi.StringInput `pulumi:"loggingConfigurationName"`
	// The name of the network security perimeter.
	NetworkSecurityPerimeterName pulumi.StringInput `pulumi:"networkSecurityPerimeterName"`
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput `pulumi:"resourceGroupName"`
}

func (LookupNetworkSecurityPerimeterLoggingConfigurationOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupNetworkSecurityPerimeterLoggingConfigurationArgs)(nil)).Elem()
}

// The NSP logging configuration
type LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput struct{ *pulumi.OutputState }

func (LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupNetworkSecurityPerimeterLoggingConfigurationResult)(nil)).Elem()
}

func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) ToLookupNetworkSecurityPerimeterLoggingConfigurationResultOutput() LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput {
	return o
}

func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) ToLookupNetworkSecurityPerimeterLoggingConfigurationResultOutputWithContext(ctx context.Context) LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput {
	return o
}

// A unique read-only string that changes whenever the resource is updated.
func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) Etag() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkSecurityPerimeterLoggingConfigurationResult) string { return v.Etag }).(pulumi.StringOutput)
}

// Resource ID.
func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkSecurityPerimeterLoggingConfigurationResult) string { return v.Id }).(pulumi.StringOutput)
}

// Resource name.
func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkSecurityPerimeterLoggingConfigurationResult) string { return v.Name }).(pulumi.StringOutput)
}

// Properties of the NSP logging configuration.
func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) Properties() NspLoggingConfigurationPropertiesResponseOutput {
	return o.ApplyT(func(v LookupNetworkSecurityPerimeterLoggingConfigurationResult) NspLoggingConfigurationPropertiesResponse {
		return v.Properties
	}).(NspLoggingConfigurationPropertiesResponseOutput)
}

// Resource type.
func (o LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkSecurityPerimeterLoggingConfigurationResult) string { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupNetworkSecurityPerimeterLoggingConfigurationResultOutput{})
}
