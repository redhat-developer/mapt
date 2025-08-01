// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package ec2

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Use this data source to get information about a Network Interface.
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
//			_, err := ec2.LookupNetworkInterface(ctx, &ec2.LookupNetworkInterfaceArgs{
//				Id: pulumi.StringRef("eni-01234567"),
//			}, nil)
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
func LookupNetworkInterface(ctx *pulumi.Context, args *LookupNetworkInterfaceArgs, opts ...pulumi.InvokeOption) (*LookupNetworkInterfaceResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv LookupNetworkInterfaceResult
	err := ctx.Invoke("aws:ec2/getNetworkInterface:getNetworkInterface", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getNetworkInterface.
type LookupNetworkInterfaceArgs struct {
	// One or more name/value pairs to filter off of. There are several valid keys, for a full reference, check out [describe-network-interfaces](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-interfaces.html) in the AWS CLI reference.
	Filters []GetNetworkInterfaceFilter `pulumi:"filters"`
	// Identifier for the network interface.
	Id *string `pulumi:"id"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region *string `pulumi:"region"`
	// Any tags assigned to the network interface.
	Tags map[string]string `pulumi:"tags"`
}

// A collection of values returned by getNetworkInterface.
type LookupNetworkInterfaceResult struct {
	// ARN of the network interface.
	Arn string `pulumi:"arn"`
	// Association information for an Elastic IP address (IPv4) associated with the network interface. See supported fields below.
	Associations []GetNetworkInterfaceAssociation    `pulumi:"associations"`
	Attachments  []GetNetworkInterfaceAttachmentType `pulumi:"attachments"`
	// Availability Zone.
	AvailabilityZone string `pulumi:"availabilityZone"`
	// Description of the network interface.
	Description string                      `pulumi:"description"`
	Filters     []GetNetworkInterfaceFilter `pulumi:"filters"`
	Id          string                      `pulumi:"id"`
	// Type of interface.
	InterfaceType string `pulumi:"interfaceType"`
	// List of IPv6 addresses to assign to the ENI.
	Ipv6Addresses []string `pulumi:"ipv6Addresses"`
	// MAC address.
	MacAddress string `pulumi:"macAddress"`
	// ARN of the Outpost.
	OutpostArn string `pulumi:"outpostArn"`
	// AWS account ID of the owner of the network interface.
	OwnerId string `pulumi:"ownerId"`
	// Private DNS name.
	PrivateDnsName string `pulumi:"privateDnsName"`
	// Private IPv4 address of the network interface within the subnet.
	PrivateIp string `pulumi:"privateIp"`
	// Private IPv4 addresses associated with the network interface.
	PrivateIps []string `pulumi:"privateIps"`
	Region     string   `pulumi:"region"`
	// ID of the entity that launched the instance on your behalf.
	RequesterId string `pulumi:"requesterId"`
	// List of security groups for the network interface.
	SecurityGroups []string `pulumi:"securityGroups"`
	// ID of the subnet.
	SubnetId string `pulumi:"subnetId"`
	// Any tags assigned to the network interface.
	Tags map[string]string `pulumi:"tags"`
	// ID of the VPC.
	VpcId string `pulumi:"vpcId"`
}

func LookupNetworkInterfaceOutput(ctx *pulumi.Context, args LookupNetworkInterfaceOutputArgs, opts ...pulumi.InvokeOption) LookupNetworkInterfaceResultOutput {
	return pulumi.ToOutputWithContext(ctx.Context(), args).
		ApplyT(func(v interface{}) (LookupNetworkInterfaceResultOutput, error) {
			args := v.(LookupNetworkInterfaceArgs)
			options := pulumi.InvokeOutputOptions{InvokeOptions: internal.PkgInvokeDefaultOpts(opts)}
			return ctx.InvokeOutput("aws:ec2/getNetworkInterface:getNetworkInterface", args, LookupNetworkInterfaceResultOutput{}, options).(LookupNetworkInterfaceResultOutput), nil
		}).(LookupNetworkInterfaceResultOutput)
}

// A collection of arguments for invoking getNetworkInterface.
type LookupNetworkInterfaceOutputArgs struct {
	// One or more name/value pairs to filter off of. There are several valid keys, for a full reference, check out [describe-network-interfaces](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-interfaces.html) in the AWS CLI reference.
	Filters GetNetworkInterfaceFilterArrayInput `pulumi:"filters"`
	// Identifier for the network interface.
	Id pulumi.StringPtrInput `pulumi:"id"`
	// Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration.
	Region pulumi.StringPtrInput `pulumi:"region"`
	// Any tags assigned to the network interface.
	Tags pulumi.StringMapInput `pulumi:"tags"`
}

func (LookupNetworkInterfaceOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupNetworkInterfaceArgs)(nil)).Elem()
}

// A collection of values returned by getNetworkInterface.
type LookupNetworkInterfaceResultOutput struct{ *pulumi.OutputState }

func (LookupNetworkInterfaceResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupNetworkInterfaceResult)(nil)).Elem()
}

func (o LookupNetworkInterfaceResultOutput) ToLookupNetworkInterfaceResultOutput() LookupNetworkInterfaceResultOutput {
	return o
}

func (o LookupNetworkInterfaceResultOutput) ToLookupNetworkInterfaceResultOutputWithContext(ctx context.Context) LookupNetworkInterfaceResultOutput {
	return o
}

// ARN of the network interface.
func (o LookupNetworkInterfaceResultOutput) Arn() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.Arn }).(pulumi.StringOutput)
}

// Association information for an Elastic IP address (IPv4) associated with the network interface. See supported fields below.
func (o LookupNetworkInterfaceResultOutput) Associations() GetNetworkInterfaceAssociationArrayOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) []GetNetworkInterfaceAssociation { return v.Associations }).(GetNetworkInterfaceAssociationArrayOutput)
}

func (o LookupNetworkInterfaceResultOutput) Attachments() GetNetworkInterfaceAttachmentTypeArrayOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) []GetNetworkInterfaceAttachmentType { return v.Attachments }).(GetNetworkInterfaceAttachmentTypeArrayOutput)
}

// Availability Zone.
func (o LookupNetworkInterfaceResultOutput) AvailabilityZone() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.AvailabilityZone }).(pulumi.StringOutput)
}

// Description of the network interface.
func (o LookupNetworkInterfaceResultOutput) Description() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.Description }).(pulumi.StringOutput)
}

func (o LookupNetworkInterfaceResultOutput) Filters() GetNetworkInterfaceFilterArrayOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) []GetNetworkInterfaceFilter { return v.Filters }).(GetNetworkInterfaceFilterArrayOutput)
}

func (o LookupNetworkInterfaceResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.Id }).(pulumi.StringOutput)
}

// Type of interface.
func (o LookupNetworkInterfaceResultOutput) InterfaceType() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.InterfaceType }).(pulumi.StringOutput)
}

// List of IPv6 addresses to assign to the ENI.
func (o LookupNetworkInterfaceResultOutput) Ipv6Addresses() pulumi.StringArrayOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) []string { return v.Ipv6Addresses }).(pulumi.StringArrayOutput)
}

// MAC address.
func (o LookupNetworkInterfaceResultOutput) MacAddress() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.MacAddress }).(pulumi.StringOutput)
}

// ARN of the Outpost.
func (o LookupNetworkInterfaceResultOutput) OutpostArn() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.OutpostArn }).(pulumi.StringOutput)
}

// AWS account ID of the owner of the network interface.
func (o LookupNetworkInterfaceResultOutput) OwnerId() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.OwnerId }).(pulumi.StringOutput)
}

// Private DNS name.
func (o LookupNetworkInterfaceResultOutput) PrivateDnsName() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.PrivateDnsName }).(pulumi.StringOutput)
}

// Private IPv4 address of the network interface within the subnet.
func (o LookupNetworkInterfaceResultOutput) PrivateIp() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.PrivateIp }).(pulumi.StringOutput)
}

// Private IPv4 addresses associated with the network interface.
func (o LookupNetworkInterfaceResultOutput) PrivateIps() pulumi.StringArrayOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) []string { return v.PrivateIps }).(pulumi.StringArrayOutput)
}

func (o LookupNetworkInterfaceResultOutput) Region() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.Region }).(pulumi.StringOutput)
}

// ID of the entity that launched the instance on your behalf.
func (o LookupNetworkInterfaceResultOutput) RequesterId() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.RequesterId }).(pulumi.StringOutput)
}

// List of security groups for the network interface.
func (o LookupNetworkInterfaceResultOutput) SecurityGroups() pulumi.StringArrayOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) []string { return v.SecurityGroups }).(pulumi.StringArrayOutput)
}

// ID of the subnet.
func (o LookupNetworkInterfaceResultOutput) SubnetId() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.SubnetId }).(pulumi.StringOutput)
}

// Any tags assigned to the network interface.
func (o LookupNetworkInterfaceResultOutput) Tags() pulumi.StringMapOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) map[string]string { return v.Tags }).(pulumi.StringMapOutput)
}

// ID of the VPC.
func (o LookupNetworkInterfaceResultOutput) VpcId() pulumi.StringOutput {
	return o.ApplyT(func(v LookupNetworkInterfaceResult) string { return v.VpcId }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupNetworkInterfaceResultOutput{})
}
