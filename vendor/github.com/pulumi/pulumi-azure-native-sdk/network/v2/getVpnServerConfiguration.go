// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package network

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Retrieves the details of a VpnServerConfiguration.
//
// Uses Azure REST API version 2023-02-01.
//
// Other available API versions: 2023-04-01, 2023-05-01, 2023-06-01, 2023-09-01, 2023-11-01, 2024-01-01, 2024-03-01, 2024-05-01.
func LookupVpnServerConfiguration(ctx *pulumi.Context, args *LookupVpnServerConfigurationArgs, opts ...pulumi.InvokeOption) (*LookupVpnServerConfigurationResult, error) {
	opts = utilities.PkgInvokeDefaultOpts(opts)
	var rv LookupVpnServerConfigurationResult
	err := ctx.Invoke("azure-native:network:getVpnServerConfiguration", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type LookupVpnServerConfigurationArgs struct {
	// The resource group name of the VpnServerConfiguration.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// The name of the VpnServerConfiguration being retrieved.
	VpnServerConfigurationName string `pulumi:"vpnServerConfigurationName"`
}

// VpnServerConfiguration Resource.
type LookupVpnServerConfigurationResult struct {
	// The set of aad vpn authentication parameters.
	AadAuthenticationParameters *AadAuthenticationParametersResponse `pulumi:"aadAuthenticationParameters"`
	// List of all VpnServerConfigurationPolicyGroups.
	ConfigurationPolicyGroups []VpnServerConfigurationPolicyGroupResponse `pulumi:"configurationPolicyGroups"`
	// A unique read-only string that changes whenever the resource is updated.
	Etag string `pulumi:"etag"`
	// Resource ID.
	Id *string `pulumi:"id"`
	// Resource location.
	Location *string `pulumi:"location"`
	// Resource name.
	Name string `pulumi:"name"`
	// List of references to P2SVpnGateways.
	P2SVpnGateways []P2SVpnGatewayResponse `pulumi:"p2SVpnGateways"`
	// The provisioning state of the VpnServerConfiguration resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState string `pulumi:"provisioningState"`
	// Radius client root certificate of VpnServerConfiguration.
	RadiusClientRootCertificates []VpnServerConfigRadiusClientRootCertificateResponse `pulumi:"radiusClientRootCertificates"`
	// The radius server address property of the VpnServerConfiguration resource for point to site client connection.
	RadiusServerAddress *string `pulumi:"radiusServerAddress"`
	// Radius Server root certificate of VpnServerConfiguration.
	RadiusServerRootCertificates []VpnServerConfigRadiusServerRootCertificateResponse `pulumi:"radiusServerRootCertificates"`
	// The radius secret property of the VpnServerConfiguration resource for point to site client connection.
	RadiusServerSecret *string `pulumi:"radiusServerSecret"`
	// Multiple Radius Server configuration for VpnServerConfiguration.
	RadiusServers []RadiusServerResponse `pulumi:"radiusServers"`
	// Resource tags.
	Tags map[string]string `pulumi:"tags"`
	// Resource type.
	Type string `pulumi:"type"`
	// VPN authentication types for the VpnServerConfiguration.
	VpnAuthenticationTypes []string `pulumi:"vpnAuthenticationTypes"`
	// VpnClientIpsecPolicies for VpnServerConfiguration.
	VpnClientIpsecPolicies []IpsecPolicyResponse `pulumi:"vpnClientIpsecPolicies"`
	// VPN client revoked certificate of VpnServerConfiguration.
	VpnClientRevokedCertificates []VpnServerConfigVpnClientRevokedCertificateResponse `pulumi:"vpnClientRevokedCertificates"`
	// VPN client root certificate of VpnServerConfiguration.
	VpnClientRootCertificates []VpnServerConfigVpnClientRootCertificateResponse `pulumi:"vpnClientRootCertificates"`
	// VPN protocols for the VpnServerConfiguration.
	VpnProtocols []string `pulumi:"vpnProtocols"`
}

func LookupVpnServerConfigurationOutput(ctx *pulumi.Context, args LookupVpnServerConfigurationOutputArgs, opts ...pulumi.InvokeOption) LookupVpnServerConfigurationResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupVpnServerConfigurationResultOutput, error) {
			args := v.(LookupVpnServerConfigurationArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: utilities.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("azure-native:network:getVpnServerConfiguration", args, LookupVpnServerConfigurationResultOutput{}, options).(LookupVpnServerConfigurationResultOutput), nil
		}).(LookupVpnServerConfigurationResultOutput)
}

type LookupVpnServerConfigurationOutputArgs struct {
	// The resource group name of the VpnServerConfiguration.
	ResourceGroupName pulumi.StringInput `pulumi:"resourceGroupName"`
	// The name of the VpnServerConfiguration being retrieved.
	VpnServerConfigurationName pulumi.StringInput `pulumi:"vpnServerConfigurationName"`
}

func (LookupVpnServerConfigurationOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupVpnServerConfigurationArgs)(nil)).Elem()
}

// VpnServerConfiguration Resource.
type LookupVpnServerConfigurationResultOutput struct{ *pulumi.OutputState }

func (LookupVpnServerConfigurationResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupVpnServerConfigurationResult)(nil)).Elem()
}

func (o LookupVpnServerConfigurationResultOutput) ToLookupVpnServerConfigurationResultOutput() LookupVpnServerConfigurationResultOutput {
	return o
}

func (o LookupVpnServerConfigurationResultOutput) ToLookupVpnServerConfigurationResultOutputWithContext(ctx context.Context) LookupVpnServerConfigurationResultOutput {
	return o
}

// The set of aad vpn authentication parameters.
func (o LookupVpnServerConfigurationResultOutput) AadAuthenticationParameters() AadAuthenticationParametersResponsePtrOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) *AadAuthenticationParametersResponse {
		return v.AadAuthenticationParameters
	}).(AadAuthenticationParametersResponsePtrOutput)
}

// List of all VpnServerConfigurationPolicyGroups.
func (o LookupVpnServerConfigurationResultOutput) ConfigurationPolicyGroups() VpnServerConfigurationPolicyGroupResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []VpnServerConfigurationPolicyGroupResponse {
		return v.ConfigurationPolicyGroups
	}).(VpnServerConfigurationPolicyGroupResponseArrayOutput)
}

// A unique read-only string that changes whenever the resource is updated.
func (o LookupVpnServerConfigurationResultOutput) Etag() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) string { return v.Etag }).(pulumi.StringOutput)
}

// Resource ID.
func (o LookupVpnServerConfigurationResultOutput) Id() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) *string { return v.Id }).(pulumi.StringPtrOutput)
}

// Resource location.
func (o LookupVpnServerConfigurationResultOutput) Location() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) *string { return v.Location }).(pulumi.StringPtrOutput)
}

// Resource name.
func (o LookupVpnServerConfigurationResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) string { return v.Name }).(pulumi.StringOutput)
}

// List of references to P2SVpnGateways.
func (o LookupVpnServerConfigurationResultOutput) P2SVpnGateways() P2SVpnGatewayResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []P2SVpnGatewayResponse { return v.P2SVpnGateways }).(P2SVpnGatewayResponseArrayOutput)
}

// The provisioning state of the VpnServerConfiguration resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
func (o LookupVpnServerConfigurationResultOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) string { return v.ProvisioningState }).(pulumi.StringOutput)
}

// Radius client root certificate of VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) RadiusClientRootCertificates() VpnServerConfigRadiusClientRootCertificateResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []VpnServerConfigRadiusClientRootCertificateResponse {
		return v.RadiusClientRootCertificates
	}).(VpnServerConfigRadiusClientRootCertificateResponseArrayOutput)
}

// The radius server address property of the VpnServerConfiguration resource for point to site client connection.
func (o LookupVpnServerConfigurationResultOutput) RadiusServerAddress() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) *string { return v.RadiusServerAddress }).(pulumi.StringPtrOutput)
}

// Radius Server root certificate of VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) RadiusServerRootCertificates() VpnServerConfigRadiusServerRootCertificateResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []VpnServerConfigRadiusServerRootCertificateResponse {
		return v.RadiusServerRootCertificates
	}).(VpnServerConfigRadiusServerRootCertificateResponseArrayOutput)
}

// The radius secret property of the VpnServerConfiguration resource for point to site client connection.
func (o LookupVpnServerConfigurationResultOutput) RadiusServerSecret() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) *string { return v.RadiusServerSecret }).(pulumi.StringPtrOutput)
}

// Multiple Radius Server configuration for VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) RadiusServers() RadiusServerResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []RadiusServerResponse { return v.RadiusServers }).(RadiusServerResponseArrayOutput)
}

// Resource tags.
func (o LookupVpnServerConfigurationResultOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) map[string]string { return v.Tags }).(pulumi.StringMapOutput)
}

// Resource type.
func (o LookupVpnServerConfigurationResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) string { return v.Type }).(pulumi.StringOutput)
}

// VPN authentication types for the VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) VpnAuthenticationTypes() pulumi.StringArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []string { return v.VpnAuthenticationTypes }).(pulumi.StringArrayOutput)
}

// VpnClientIpsecPolicies for VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) VpnClientIpsecPolicies() IpsecPolicyResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []IpsecPolicyResponse { return v.VpnClientIpsecPolicies }).(IpsecPolicyResponseArrayOutput)
}

// VPN client revoked certificate of VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) VpnClientRevokedCertificates() VpnServerConfigVpnClientRevokedCertificateResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []VpnServerConfigVpnClientRevokedCertificateResponse {
		return v.VpnClientRevokedCertificates
	}).(VpnServerConfigVpnClientRevokedCertificateResponseArrayOutput)
}

// VPN client root certificate of VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) VpnClientRootCertificates() VpnServerConfigVpnClientRootCertificateResponseArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []VpnServerConfigVpnClientRootCertificateResponse {
		return v.VpnClientRootCertificates
	}).(VpnServerConfigVpnClientRootCertificateResponseArrayOutput)
}

// VPN protocols for the VpnServerConfiguration.
func (o LookupVpnServerConfigurationResultOutput) VpnProtocols() pulumi.StringArrayOutput {
	return o.ApplyT(func(v LookupVpnServerConfigurationResult) []string { return v.VpnProtocols }).(pulumi.StringArrayOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupVpnServerConfigurationResultOutput{})
}
