// Code generated by the Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package containerservice

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-azure-native-sdk/v2/utilities"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The configurations regarding multiple standard load balancers. If not supplied, single load balancer mode will be used. Multiple standard load balancers mode will be used if at lease one configuration is supplied. There has to be a configuration named `kubernetes`.
// Azure REST API version: 2024-03-02-preview.
//
// Other available API versions: 2024-04-02-preview, 2024-05-02-preview.
func LookupLoadBalancer(ctx *pulumi.Context, args *LookupLoadBalancerArgs, opts ...pulumi.InvokeOption) (*LookupLoadBalancerResult, error) {
	opts = utilities.PkgInvokeDefaultOpts(opts)
	var rv LookupLoadBalancerResult
	err := ctx.Invoke("azure-native:containerservice:getLoadBalancer", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type LookupLoadBalancerArgs struct {
	// The name of the load balancer.
	LoadBalancerName string `pulumi:"loadBalancerName"`
	// The name of the resource group. The name is case insensitive.
	ResourceGroupName string `pulumi:"resourceGroupName"`
	// The name of the managed cluster resource.
	ResourceName string `pulumi:"resourceName"`
}

// The configurations regarding multiple standard load balancers. If not supplied, single load balancer mode will be used. Multiple standard load balancers mode will be used if at lease one configuration is supplied. There has to be a configuration named `kubernetes`.
type LookupLoadBalancerResult struct {
	// Whether to automatically place services on the load balancer. If not supplied, the default value is true. If set to false manually, both of the external and the internal load balancer will not be selected for services unless they explicitly target it.
	AllowServicePlacement *bool `pulumi:"allowServicePlacement"`
	// Fully qualified resource ID for the resource. E.g. "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}/{resourceName}"
	Id string `pulumi:"id"`
	// The name of the resource
	Name string `pulumi:"name"`
	// Nodes that match this selector will be possible members of this load balancer.
	NodeSelector *LabelSelectorResponse `pulumi:"nodeSelector"`
	// Required field. A string value that must specify the ID of an existing agent pool. All nodes in the given pool will always be added to this load balancer. This agent pool must have at least one node and minCount>=1 for autoscaling operations. An agent pool can only be the primary pool for a single load balancer.
	PrimaryAgentPoolName string `pulumi:"primaryAgentPoolName"`
	// The current provisioning state.
	ProvisioningState string `pulumi:"provisioningState"`
	// Only services that must match this selector can be placed on this load balancer.
	ServiceLabelSelector *LabelSelectorResponse `pulumi:"serviceLabelSelector"`
	// Services created in namespaces that match the selector can be placed on this load balancer.
	ServiceNamespaceSelector *LabelSelectorResponse `pulumi:"serviceNamespaceSelector"`
	// Azure Resource Manager metadata containing createdBy and modifiedBy information.
	SystemData SystemDataResponse `pulumi:"systemData"`
	// The type of the resource. E.g. "Microsoft.Compute/virtualMachines" or "Microsoft.Storage/storageAccounts"
	Type string `pulumi:"type"`
}

func LookupLoadBalancerOutput(ctx *pulumi.Context, args LookupLoadBalancerOutputArgs, opts ...pulumi.InvokeOption) LookupLoadBalancerResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (LookupLoadBalancerResult, error) {
			args := v.(LookupLoadBalancerArgs)
			r, err := LookupLoadBalancer(ctx, &args, opts...)
			var s LookupLoadBalancerResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(LookupLoadBalancerResultOutput)
}

type LookupLoadBalancerOutputArgs struct {
	// The name of the load balancer.
	LoadBalancerName pulumi.StringInput `pulumi:"loadBalancerName"`
	// The name of the resource group. The name is case insensitive.
	ResourceGroupName pulumi.StringInput `pulumi:"resourceGroupName"`
	// The name of the managed cluster resource.
	ResourceName pulumi.StringInput `pulumi:"resourceName"`
}

func (LookupLoadBalancerOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupLoadBalancerArgs)(nil)).Elem()
}

// The configurations regarding multiple standard load balancers. If not supplied, single load balancer mode will be used. Multiple standard load balancers mode will be used if at lease one configuration is supplied. There has to be a configuration named `kubernetes`.
type LookupLoadBalancerResultOutput struct{ *pulumi.OutputState }

func (LookupLoadBalancerResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupLoadBalancerResult)(nil)).Elem()
}

func (o LookupLoadBalancerResultOutput) ToLookupLoadBalancerResultOutput() LookupLoadBalancerResultOutput {
	return o
}

func (o LookupLoadBalancerResultOutput) ToLookupLoadBalancerResultOutputWithContext(ctx context.Context) LookupLoadBalancerResultOutput {
	return o
}

// Whether to automatically place services on the load balancer. If not supplied, the default value is true. If set to false manually, both of the external and the internal load balancer will not be selected for services unless they explicitly target it.
func (o LookupLoadBalancerResultOutput) AllowServicePlacement() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) *bool { return v.AllowServicePlacement }).(pulumi.BoolPtrOutput)
}

// Fully qualified resource ID for the resource. E.g. "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}/{resourceName}"
func (o LookupLoadBalancerResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) string { return v.Id }).(pulumi.StringOutput)
}

// The name of the resource
func (o LookupLoadBalancerResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) string { return v.Name }).(pulumi.StringOutput)
}

// Nodes that match this selector will be possible members of this load balancer.
func (o LookupLoadBalancerResultOutput) NodeSelector() LabelSelectorResponsePtrOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) *LabelSelectorResponse { return v.NodeSelector }).(LabelSelectorResponsePtrOutput)
}

// Required field. A string value that must specify the ID of an existing agent pool. All nodes in the given pool will always be added to this load balancer. This agent pool must have at least one node and minCount>=1 for autoscaling operations. An agent pool can only be the primary pool for a single load balancer.
func (o LookupLoadBalancerResultOutput) PrimaryAgentPoolName() pulumi.StringOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) string { return v.PrimaryAgentPoolName }).(pulumi.StringOutput)
}

// The current provisioning state.
func (o LookupLoadBalancerResultOutput) ProvisioningState() pulumi.StringOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) string { return v.ProvisioningState }).(pulumi.StringOutput)
}

// Only services that must match this selector can be placed on this load balancer.
func (o LookupLoadBalancerResultOutput) ServiceLabelSelector() LabelSelectorResponsePtrOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) *LabelSelectorResponse { return v.ServiceLabelSelector }).(LabelSelectorResponsePtrOutput)
}

// Services created in namespaces that match the selector can be placed on this load balancer.
func (o LookupLoadBalancerResultOutput) ServiceNamespaceSelector() LabelSelectorResponsePtrOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) *LabelSelectorResponse { return v.ServiceNamespaceSelector }).(LabelSelectorResponsePtrOutput)
}

// Azure Resource Manager metadata containing createdBy and modifiedBy information.
func (o LookupLoadBalancerResultOutput) SystemData() SystemDataResponseOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) SystemDataResponse { return v.SystemData }).(SystemDataResponseOutput)
}

// The type of the resource. E.g. "Microsoft.Compute/virtualMachines" or "Microsoft.Storage/storageAccounts"
func (o LookupLoadBalancerResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupLoadBalancerResult) string { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupLoadBalancerResultOutput{})
}