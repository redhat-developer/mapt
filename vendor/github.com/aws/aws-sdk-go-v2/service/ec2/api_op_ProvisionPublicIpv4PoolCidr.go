// Code generated by smithy-go-codegen DO NOT EDIT.

package ec2

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Provision a CIDR to a public IPv4 pool.
//
// For more information about IPAM, see [What is IPAM?] in the Amazon VPC IPAM User Guide.
//
// [What is IPAM?]: https://docs.aws.amazon.com/vpc/latest/ipam/what-is-it-ipam.html
func (c *Client) ProvisionPublicIpv4PoolCidr(ctx context.Context, params *ProvisionPublicIpv4PoolCidrInput, optFns ...func(*Options)) (*ProvisionPublicIpv4PoolCidrOutput, error) {
	if params == nil {
		params = &ProvisionPublicIpv4PoolCidrInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ProvisionPublicIpv4PoolCidr", params, optFns, c.addOperationProvisionPublicIpv4PoolCidrMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ProvisionPublicIpv4PoolCidrOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ProvisionPublicIpv4PoolCidrInput struct {

	// The ID of the IPAM pool you would like to use to allocate this CIDR.
	//
	// This member is required.
	IpamPoolId *string

	// The netmask length of the CIDR you would like to allocate to the public IPv4
	// pool. The least specific netmask length you can define is 24.
	//
	// This member is required.
	NetmaskLength *int32

	// The ID of the public IPv4 pool you would like to use for this CIDR.
	//
	// This member is required.
	PoolId *string

	// A check for whether you have the required permissions for the action without
	// actually making the request and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// The Availability Zone (AZ) or Local Zone (LZ) network border group that the
	// resource that the IP address is assigned to is in. Defaults to an AZ network
	// border group. For more information on available Local Zones, see [Local Zone availability]in the Amazon
	// EC2 User Guide.
	//
	// [Local Zone availability]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-byoip.html#byoip-zone-avail
	NetworkBorderGroup *string

	noSmithyDocumentSerde
}

type ProvisionPublicIpv4PoolCidrOutput struct {

	// Information about the address range of the public IPv4 pool.
	PoolAddressRange *types.PublicIpv4PoolRange

	// The ID of the pool that you want to provision the CIDR to.
	PoolId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationProvisionPublicIpv4PoolCidrMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsEc2query_serializeOpProvisionPublicIpv4PoolCidr{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpProvisionPublicIpv4PoolCidr{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "ProvisionPublicIpv4PoolCidr"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addComputePayloadSHA256(stack); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addSpanRetryLoop(stack, options); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addTimeOffsetBuild(stack, c); err != nil {
		return err
	}
	if err = addUserAgentRetryMode(stack, options); err != nil {
		return err
	}
	if err = addCredentialSource(stack, options); err != nil {
		return err
	}
	if err = addOpProvisionPublicIpv4PoolCidrValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opProvisionPublicIpv4PoolCidr(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	if err = addInterceptBeforeRetryLoop(stack, options); err != nil {
		return err
	}
	if err = addInterceptAttempt(stack, options); err != nil {
		return err
	}
	if err = addInterceptExecution(stack, options); err != nil {
		return err
	}
	if err = addInterceptBeforeSerialization(stack, options); err != nil {
		return err
	}
	if err = addInterceptAfterSerialization(stack, options); err != nil {
		return err
	}
	if err = addInterceptBeforeSigning(stack, options); err != nil {
		return err
	}
	if err = addInterceptAfterSigning(stack, options); err != nil {
		return err
	}
	if err = addInterceptTransmit(stack, options); err != nil {
		return err
	}
	if err = addInterceptBeforeDeserialization(stack, options); err != nil {
		return err
	}
	if err = addInterceptAfterDeserialization(stack, options); err != nil {
		return err
	}
	if err = addSpanInitializeStart(stack); err != nil {
		return err
	}
	if err = addSpanInitializeEnd(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestStart(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestEnd(stack); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opProvisionPublicIpv4PoolCidr(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "ProvisionPublicIpv4PoolCidr",
	}
}
