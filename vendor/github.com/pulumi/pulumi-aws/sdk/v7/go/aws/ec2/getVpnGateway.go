// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package ec2

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The VPN Gateway data source provides details about
// a specific VPN gateway.
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			selected, err := ec2.LookupVpnGateway(ctx, &ec2.LookupVpnGatewayArgs{
//				Filters: []ec2.GetVpnGatewayFilter{
//					{
//						Name: "tag:Name",
//						Values: []string{
//							"vpn-gw",
//						},
//					},
//				},
//			}, nil)
//			if err != nil {
//				return err
//			}
//			ctx.Export("vpnGatewayId", selected.Id)
//			return nil
//		})
//	}
//
// ```
func LookupVpnGateway(ctx *pulumi.Context, args *LookupVpnGatewayArgs, opts ...pulumi.InvokeOption) (*LookupVpnGatewayResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv LookupVpnGatewayResult
	err := ctx.Invoke("aws:ec2/getVpnGateway:getVpnGateway", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getVpnGateway.
type LookupVpnGatewayArgs struct {
	// Autonomous System Number (ASN) for the Amazon side of the specific VPN Gateway to retrieve.
	//
	// The arguments of this data source act as filters for querying the available VPN gateways.
	// The given filters must match exactly one VPN gateway whose data will be exported as attributes.
	AmazonSideAsn *string `pulumi:"amazonSideAsn"`
	// ID of a VPC attached to the specific VPN Gateway to retrieve.
	AttachedVpcId *string `pulumi:"attachedVpcId"`
	// Availability Zone of the specific VPN Gateway to retrieve.
	AvailabilityZone *string `pulumi:"availabilityZone"`
	// Custom filter block as described below.
	Filters []GetVpnGatewayFilter `pulumi:"filters"`
	// ID of the specific VPN Gateway to retrieve.
	Id *string `pulumi:"id"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// State of the specific VPN Gateway to retrieve.
	State *string `pulumi:"state"`
	// Map of tags, each pair of which must exactly match
	// a pair on the desired VPN Gateway.
	Tags map[string]string `pulumi:"tags"`
}

// A collection of values returned by getVpnGateway.
type LookupVpnGatewayResult struct {
	AmazonSideAsn    string                `pulumi:"amazonSideAsn"`
	Arn              string                `pulumi:"arn"`
	AttachedVpcId    string                `pulumi:"attachedVpcId"`
	AvailabilityZone string                `pulumi:"availabilityZone"`
	Filters          []GetVpnGatewayFilter `pulumi:"filters"`
	Id               string                `pulumi:"id"`
	Region           string                `pulumi:"region"`
	State            string                `pulumi:"state"`
	Tags             map[string]string     `pulumi:"tags"`
}

func LookupVpnGatewayOutput(ctx *pulumi.Context, args LookupVpnGatewayOutputArgs, opts ...pulumi.InvokeOption) LookupVpnGatewayResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupVpnGatewayResultOutput, error) {
			args := v.(LookupVpnGatewayArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:ec2/getVpnGateway:getVpnGateway", args, LookupVpnGatewayResultOutput{}, options).(LookupVpnGatewayResultOutput), nil
		}).(LookupVpnGatewayResultOutput)
}

// A collection of arguments for invoking getVpnGateway.
type LookupVpnGatewayOutputArgs struct {
	// Autonomous System Number (ASN) for the Amazon side of the specific VPN Gateway to retrieve.
	//
	// The arguments of this data source act as filters for querying the available VPN gateways.
	// The given filters must match exactly one VPN gateway whose data will be exported as attributes.
	AmazonSideAsn pulumi.StringPtrInput `pulumi:"amazonSideAsn"`
	// ID of a VPC attached to the specific VPN Gateway to retrieve.
	AttachedVpcId pulumi.StringPtrInput `pulumi:"attachedVpcId"`
	// Availability Zone of the specific VPN Gateway to retrieve.
	AvailabilityZone pulumi.StringPtrInput `pulumi:"availabilityZone"`
	// Custom filter block as described below.
	Filters GetVpnGatewayFilterArrayInput `pulumi:"filters"`
	// ID of the specific VPN Gateway to retrieve.
	Id pulumi.StringPtrInput `pulumi:"id"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput `pulumi:"region"`
	// State of the specific VPN Gateway to retrieve.
	State pulumi.StringPtrInput `pulumi:"state"`
	// Map of tags, each pair of which must exactly match
	// a pair on the desired VPN Gateway.
	Tags pulumi.StringMapInput `pulumi:"tags"`
}

func (LookupVpnGatewayOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupVpnGatewayArgs)(nil)).Elem()
}

// A collection of values returned by getVpnGateway.
type LookupVpnGatewayResultOutput struct{ *pulumi.OutputState }

func (LookupVpnGatewayResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupVpnGatewayResult)(nil)).Elem()
}

func (o LookupVpnGatewayResultOutput) ToLookupVpnGatewayResultOutput() LookupVpnGatewayResultOutput {
	return o
}

func (o LookupVpnGatewayResultOutput) ToLookupVpnGatewayResultOutputWithContext(ctx context.Context) LookupVpnGatewayResultOutput {
	return o
}

func (o LookupVpnGatewayResultOutput) AmazonSideAsn() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.AmazonSideAsn }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) Arn() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.Arn }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) AttachedVpcId() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.AttachedVpcId }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) AvailabilityZone() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.AvailabilityZone }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) Filters() GetVpnGatewayFilterArrayOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) []GetVpnGatewayFilter { return v.Filters }).(GetVpnGatewayFilterArrayOutput)
}

func (o LookupVpnGatewayResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.Id }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.Region }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) State() pulumi.StringOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) string { return v.State }).(pulumi.StringOutput)
}

func (o LookupVpnGatewayResultOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v LookupVpnGatewayResult) map[string]string { return v.Tags }).(pulumi.StringMapOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupVpnGatewayResultOutput{})
}
