// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package compute

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Retrieves information about a gallery Application Definition.
//
// Uses Azure REST API version 2022-03-03.
//
// Other available API versions: 2022-08-03, 2023-07-03, 2024-03-03.
func LookupGalleryApplication(ctx *pulumi.Context, args *LookupGalleryApplicationArgs, opts ...pulumi.InvokeOption) (*LookupGalleryApplicationResult, error) {
	opts = utilities.PkgInvokeDefaultOpts(opts)
	var rv LookupGalleryApplicationResult
	err := ctx.Invoke("azure-native:compute:getGalleryApplication", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type LookupGalleryApplicationArgs struct {
	// The name of the gallery Application Definition to be retrieved.
	GalleryApplicationName string `pulumi:"galleryApplicationName"`
	// The name of the Shared Application Gallery from which the Application Definitions are to be retrieved.
	GalleryName string `pulumi:"galleryName"`
	// The name of the resource group.
	ResourceGroupName string `pulumi:"resourceGroupName"`
}

// Specifies information about the gallery Application Definition that you want to create or update.
type LookupGalleryApplicationResult struct {
	// A list of custom actions that can be performed with all of the Gallery Application Versions within this Gallery Application.
	CustomActions []GalleryApplicationCustomActionResponse `pulumi:"customActions"`
	// The description of this gallery Application Definition resource. This property is updatable.
	Description *string `pulumi:"description"`
	// The end of life date of the gallery Application Definition. This property can be used for decommissioning purposes. This property is updatable.
	EndOfLifeDate *string `pulumi:"endOfLifeDate"`
	// The Eula agreement for the gallery Application Definition.
	Eula *string `pulumi:"eula"`
	// Resource Id
	Id string `pulumi:"id"`
	// Resource location
	Location string `pulumi:"location"`
	// Resource name
	Name string `pulumi:"name"`
	// The privacy statement uri.
	PrivacyStatementUri *string `pulumi:"privacyStatementUri"`
	// The release note uri.
	ReleaseNoteUri *string `pulumi:"releaseNoteUri"`
	// This property allows you to specify the supported type of the OS that application is built for. <br><br> Possible values are: <br><br> **Windows** <br><br> **Linux**
	SupportedOSType string `pulumi:"supportedOSType"`
	// Resource tags
	Tags map[string]string `pulumi:"tags"`
	// Resource type
	Type string `pulumi:"type"`
}

func LookupGalleryApplicationOutput(ctx *pulumi.Context, args LookupGalleryApplicationOutputArgs, opts ...pulumi.InvokeOption) LookupGalleryApplicationResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupGalleryApplicationResultOutput, error) {
			args := v.(LookupGalleryApplicationArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: utilities.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("azure-native:compute:getGalleryApplication", args, LookupGalleryApplicationResultOutput{}, options).(LookupGalleryApplicationResultOutput), nil
		}).(LookupGalleryApplicationResultOutput)
}

type LookupGalleryApplicationOutputArgs struct {
	// The name of the gallery Application Definition to be retrieved.
	GalleryApplicationName pulumi.StringInput `pulumi:"galleryApplicationName"`
	// The name of the Shared Application Gallery from which the Application Definitions are to be retrieved.
	GalleryName pulumi.StringInput `pulumi:"galleryName"`
	// The name of the resource group.
	ResourceGroupName pulumi.StringInput `pulumi:"resourceGroupName"`
}

func (LookupGalleryApplicationOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupGalleryApplicationArgs)(nil)).Elem()
}

// Specifies information about the gallery Application Definition that you want to create or update.
type LookupGalleryApplicationResultOutput struct{ *pulumi.OutputState }

func (LookupGalleryApplicationResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupGalleryApplicationResult)(nil)).Elem()
}

func (o LookupGalleryApplicationResultOutput) ToLookupGalleryApplicationResultOutput() LookupGalleryApplicationResultOutput {
	return o
}

func (o LookupGalleryApplicationResultOutput) ToLookupGalleryApplicationResultOutputWithContext(ctx context.Context) LookupGalleryApplicationResultOutput {
	return o
}

// A list of custom actions that can be performed with all of the Gallery Application Versions within this Gallery Application.
func (o LookupGalleryApplicationResultOutput) CustomActions() GalleryApplicationCustomActionResponseArrayOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) []GalleryApplicationCustomActionResponse {
		return v.CustomActions
	}).(GalleryApplicationCustomActionResponseArrayOutput)
}

// The description of this gallery Application Definition resource. This property is updatable.
func (o LookupGalleryApplicationResultOutput) Description() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) *string { return v.Description }).(pulumi.StringPtrOutput)
}

// The end of life date of the gallery Application Definition. This property can be used for decommissioning purposes. This property is updatable.
func (o LookupGalleryApplicationResultOutput) EndOfLifeDate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) *string { return v.EndOfLifeDate }).(pulumi.StringPtrOutput)
}

// The Eula agreement for the gallery Application Definition.
func (o LookupGalleryApplicationResultOutput) Eula() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) *string { return v.Eula }).(pulumi.StringPtrOutput)
}

// Resource Id
func (o LookupGalleryApplicationResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) string { return v.Id }).(pulumi.StringOutput)
}

// Resource location
func (o LookupGalleryApplicationResultOutput) Location() pulumi.StringOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) string { return v.Location }).(pulumi.StringOutput)
}

// Resource name
func (o LookupGalleryApplicationResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) string { return v.Name }).(pulumi.StringOutput)
}

// The privacy statement uri.
func (o LookupGalleryApplicationResultOutput) PrivacyStatementUri() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) *string { return v.PrivacyStatementUri }).(pulumi.StringPtrOutput)
}

// The release note uri.
func (o LookupGalleryApplicationResultOutput) ReleaseNoteUri() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) *string { return v.ReleaseNoteUri }).(pulumi.StringPtrOutput)
}

// This property allows you to specify the supported type of the OS that application is built for. <br><br> Possible values are: <br><br> **Windows** <br><br> **Linux**
func (o LookupGalleryApplicationResultOutput) SupportedOSType() pulumi.StringOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) string { return v.SupportedOSType }).(pulumi.StringOutput)
}

// Resource tags
func (o LookupGalleryApplicationResultOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) map[string]string { return v.Tags }).(pulumi.StringMapOutput)
}

// Resource type
func (o LookupGalleryApplicationResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupGalleryApplicationResult) string { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupGalleryApplicationResultOutput{})
}
