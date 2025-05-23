// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package compute

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The operation to get the extension.
//
// Uses Azure REST API version 2023-03-01.
//
// Other available API versions: 2021-11-01, 2023-07-01, 2023-09-01, 2024-03-01, 2024-07-01, 2024-11-01.
func LookupVirtualMachineExtension(ctx *pulumi.Context, args *LookupVirtualMachineExtensionArgs, opts ...pulumi.InvokeOption) (*LookupVirtualMachineExtensionResult, error) {
	opts = utilities.PkgInvokeDefaultOpts(opts)
	var rv LookupVirtualMachineExtensionResult
	err := ctx.Invoke("azure-native:compute:getVirtualMachineExtension", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type LookupVirtualMachineExtensionArgs struct {
	// The expand expression to apply on the operation.
	Expand *string `pulumi:"expand"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// The name of the virtual machine extension.
	VmExtensionName string `pulumi:"vmExtensionName"`
	// The name of the virtual machine containing the extension.
	VmName string `pulumi:"vmName"`
}

// Describes a Virtual Machine Extension.
type LookupVirtualMachineExtensionResult struct {
	// Indicates whether the extension should use a newer minor version if one is available at deployment time. Once deployed, however, the extension will not upgrade minor versions unless redeployed, even with this property set to true.
	AutoUpgradeMinorVersion *bool `pulumi:"autoUpgradeMinorVersion"`
	// Indicates whether the extension should be automatically upgraded by the platform if there is a newer version of the extension available.
	EnableAutomaticUpgrade *bool `pulumi:"enableAutomaticUpgrade"`
	// How the extension handler should be forced to update even if the extension configuration has not changed.
	ForceUpdateTag *string `pulumi:"forceUpdateTag"`
	// Resource Id
	Id string `pulumi:"id"`
	// The virtual machine extension instance view.
	InstanceView *VirtualMachineExtensionInstanceViewResponse `pulumi:"instanceView"`
	// Resource location
	Location *string `pulumi:"location"`
	// Resource name
	Name string `pulumi:"name"`
	// The extension can contain either protectedSettings or protectedSettingsFromKeyVault or no protected settings at all.
	ProtectedSettings interface{} `pulumi:"protectedSettings"`
	// The extensions protected settings that are passed by reference, and consumed from key vault
	ProtectedSettingsFromKeyVault *KeyVaultSecretReferenceResponse `pulumi:"protectedSettingsFromKeyVault"`
	// Collection of extension names after which this extension needs to be provisioned.
	ProvisionAfterExtensions []string `pulumi:"provisionAfterExtensions"`
	// The provisioning state, which only appears in the response.
	ProvisioningState string `pulumi:"provisioningState"`
	// The name of the extension handler publisher.
	Publisher *string `pulumi:"publisher"`
	// Json formatted public settings for the extension.
	Settings interface{} `pulumi:"settings"`
	// Indicates whether failures stemming from the extension will be suppressed (Operational failures such as not connecting to the VM will not be suppressed regardless of this value). The default is false.
	SuppressFailures *bool `pulumi:"suppressFailures"`
	// Resource tags
	Tags map[string]string `pulumi:"tags"`
	// Resource type
	Type string `pulumi:"type"`
	// Specifies the version of the script handler.
	TypeHandlerVersion *string `pulumi:"typeHandlerVersion"`
}

func LookupVirtualMachineExtensionOutput(ctx *pulumi.Context, args LookupVirtualMachineExtensionOutputArgs, opts ...pulumi.InvokeOption) LookupVirtualMachineExtensionResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupVirtualMachineExtensionResultOutput, error) {
			args := v.(LookupVirtualMachineExtensionArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: utilities.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("azure-native:compute:getVirtualMachineExtension", args, LookupVirtualMachineExtensionResultOutput{}, options).(LookupVirtualMachineExtensionResultOutput), nil
		}).(LookupVirtualMachineExtensionResultOutput)
}

type LookupVirtualMachineExtensionOutputArgs struct {
	// The expand expression to apply on the operation.
	Expand pulumi.StringPtrInput `pulumi:"expand"`
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput `pulumi:"resourceGroupName"`
	// The name of the virtual machine extension.
	VmExtensionName pulumi.StringInput `pulumi:"vmExtensionName"`
	// The name of the virtual machine containing the extension.
	VmName pulumi.StringInput `pulumi:"vmName"`
}

func (LookupVirtualMachineExtensionOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupVirtualMachineExtensionArgs)(nil)).Elem()
}

// Describes a Virtual Machine Extension.
type LookupVirtualMachineExtensionResultOutput struct{ *pulumi.OutputState }

func (LookupVirtualMachineExtensionResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupVirtualMachineExtensionResult)(nil)).Elem()
}

func (o LookupVirtualMachineExtensionResultOutput) ToLookupVirtualMachineExtensionResultOutput() LookupVirtualMachineExtensionResultOutput {
	return o
}

func (o LookupVirtualMachineExtensionResultOutput) ToLookupVirtualMachineExtensionResultOutputWithContext(ctx context.Context) LookupVirtualMachineExtensionResultOutput {
	return o
}

// Indicates whether the extension should use a newer minor version if one is available at deployment time. Once deployed, however, the extension will not upgrade minor versions unless redeployed, even with this property set to true.
func (o LookupVirtualMachineExtensionResultOutput) AutoUpgradeMinorVersion() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *bool { return v.AutoUpgradeMinorVersion }).(pulumi.BoolPtrOutput)
}

// Indicates whether the extension should be automatically upgraded by the platform if there is a newer version of the extension available.
func (o LookupVirtualMachineExtensionResultOutput) EnableAutomaticUpgrade() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *bool { return v.EnableAutomaticUpgrade }).(pulumi.BoolPtrOutput)
}

// How the extension handler should be forced to update even if the extension configuration has not changed.
func (o LookupVirtualMachineExtensionResultOutput) ForceUpdateTag() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *string { return v.ForceUpdateTag }).(pulumi.StringPtrOutput)
}

// Resource Id
func (o LookupVirtualMachineExtensionResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) string { return v.Id }).(pulumi.StringOutput)
}

// The virtual machine extension instance view.
func (o LookupVirtualMachineExtensionResultOutput) InstanceView() VirtualMachineExtensionInstanceViewResponsePtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *VirtualMachineExtensionInstanceViewResponse {
		return v.InstanceView
	}).(VirtualMachineExtensionInstanceViewResponsePtrOutput)
}

// Resource location
func (o LookupVirtualMachineExtensionResultOutput) Location() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *string { return v.Location }).(pulumi.StringPtrOutput)
}

// Resource name
func (o LookupVirtualMachineExtensionResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) string { return v.Name }).(pulumi.StringOutput)
}

// The extension can contain either protectedSettings or protectedSettingsFromKeyVault or no protected settings at all.
func (o LookupVirtualMachineExtensionResultOutput) ProtectedSettings() pulumi.AnyOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) interface{} { return v.ProtectedSettings }).(pulumi.AnyOutput)
}

// The extensions protected settings that are passed by reference, and consumed from key vault
func (o LookupVirtualMachineExtensionResultOutput) ProtectedSettingsFromKeyVault() KeyVaultSecretReferenceResponsePtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *KeyVaultSecretReferenceResponse {
		return v.ProtectedSettingsFromKeyVault
	}).(KeyVaultSecretReferenceResponsePtrOutput)
}

// Collection of extension names after which this extension needs to be provisioned.
func (o LookupVirtualMachineExtensionResultOutput) ProvisionAfterExtensions() pulumi.StringArrayOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) []string { return v.ProvisionAfterExtensions }).(pulumi.StringArrayOutput)
}

// The provisioning state, which only appears in the response.
func (o LookupVirtualMachineExtensionResultOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) string { return v.ProvisioningState }).(pulumi.StringOutput)
}

// The name of the extension handler publisher.
func (o LookupVirtualMachineExtensionResultOutput) Publisher() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *string { return v.Publisher }).(pulumi.StringPtrOutput)
}

// Json formatted public settings for the extension.
func (o LookupVirtualMachineExtensionResultOutput) Settings() pulumi.AnyOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) interface{} { return v.Settings }).(pulumi.AnyOutput)
}

// Indicates whether failures stemming from the extension will be suppressed (Operational failures such as not connecting to the VM will not be suppressed regardless of this value). The default is false.
func (o LookupVirtualMachineExtensionResultOutput) SuppressFailures() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *bool { return v.SuppressFailures }).(pulumi.BoolPtrOutput)
}

// Resource tags
func (o LookupVirtualMachineExtensionResultOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) map[string]string { return v.Tags }).(pulumi.StringMapOutput)
}

// Resource type
func (o LookupVirtualMachineExtensionResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) string { return v.Type }).(pulumi.StringOutput)
}

// Specifies the version of the script handler.
func (o LookupVirtualMachineExtensionResultOutput) TypeHandlerVersion() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVirtualMachineExtensionResult) *string { return v.TypeHandlerVersion }).(pulumi.StringPtrOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupVirtualMachineExtensionResultOutput{})
}
