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

// Allocate a CIDR from an IPAM pool. The Region you use should be the IPAM pool
// locale. The locale is the Amazon Web Services Region where this IPAM pool is
// available for allocations.
//
// In IPAM, an allocation is a CIDR assignment from an IPAM pool to another IPAM
// pool or to a resource. For more information, see [Allocate CIDRs]in the Amazon VPC IPAM User
// Guide.
//
// This action creates an allocation with strong consistency. The returned CIDR
// will not overlap with any other allocations from the same pool.
//
// [Allocate CIDRs]: https://docs.aws.amazon.com/vpc/latest/ipam/allocate-cidrs-ipam.html
func (c *Client) AllocateIpamPoolCidr(ctx context.Context, params *AllocateIpamPoolCidrInput, optFns ...func(*Options)) (*AllocateIpamPoolCidrOutput, error) {
	if params == nil {
		params = &AllocateIpamPoolCidrInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "AllocateIpamPoolCidr", params, optFns, c.addOperationAllocateIpamPoolCidrMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*AllocateIpamPoolCidrOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type AllocateIpamPoolCidrInput struct {

	// The ID of the IPAM pool from which you would like to allocate a CIDR.
	//
	// This member is required.
	IpamPoolId *string

	// Include a particular CIDR range that can be returned by the pool. Allowed CIDRs
	// are only allowed if using netmask length for allocation.
	AllowedCidrs []string

	// The CIDR you would like to allocate from the IPAM pool. Note the following:
	//
	//   - If there is no DefaultNetmaskLength allocation rule set on the pool, you
	//   must specify either the NetmaskLength or the CIDR.
	//
	//   - If the DefaultNetmaskLength allocation rule is set on the pool, you can
	//   specify either the NetmaskLength or the CIDR and the DefaultNetmaskLength
	//   allocation rule will be ignored.
	//
	// Possible values: Any available IPv4 or IPv6 CIDR.
	Cidr *string

	// A unique, case-sensitive identifier that you provide to ensure the idempotency
	// of the request. For more information, see [Ensuring idempotency].
	//
	// [Ensuring idempotency]: https://docs.aws.amazon.com/ec2/latest/devguide/ec2-api-idempotency.html
	ClientToken *string

	// A description for the allocation.
	Description *string

	// Exclude a particular CIDR range from being returned by the pool. Disallowed
	// CIDRs are only allowed if using netmask length for allocation.
	DisallowedCidrs []string

	// A check for whether you have the required permissions for the action without
	// actually making the request and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// The netmask length of the CIDR you would like to allocate from the IPAM pool.
	// Note the following:
	//
	//   - If there is no DefaultNetmaskLength allocation rule set on the pool, you
	//   must specify either the NetmaskLength or the CIDR.
	//
	//   - If the DefaultNetmaskLength allocation rule is set on the pool, you can
	//   specify either the NetmaskLength or the CIDR and the DefaultNetmaskLength
	//   allocation rule will be ignored.
	//
	// Possible netmask lengths for IPv4 addresses are 0 - 32. Possible netmask
	// lengths for IPv6 addresses are 0 - 128.
	NetmaskLength *int32

	// A preview of the next available CIDR in a pool.
	PreviewNextCidr *bool

	noSmithyDocumentSerde
}

type AllocateIpamPoolCidrOutput struct {

	// Information about the allocation created.
	IpamPoolAllocation *types.IpamPoolAllocation

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationAllocateIpamPoolCidrMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsEc2query_serializeOpAllocateIpamPoolCidr{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpAllocateIpamPoolCidr{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "AllocateIpamPoolCidr"); err != nil {
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
	if err = addIdempotencyToken_opAllocateIpamPoolCidrMiddleware(stack, options); err != nil {
		return err
	}
	if err = addOpAllocateIpamPoolCidrValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opAllocateIpamPoolCidr(options.Region), middleware.Before); err != nil {
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

type idempotencyToken_initializeOpAllocateIpamPoolCidr struct {
	tokenProvider IdempotencyTokenProvider
}

func (*idempotencyToken_initializeOpAllocateIpamPoolCidr) ID() string {
	return "OperationIdempotencyTokenAutoFill"
}

func (m *idempotencyToken_initializeOpAllocateIpamPoolCidr) HandleInitialize(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	if m.tokenProvider == nil {
		return next.HandleInitialize(ctx, in)
	}

	input, ok := in.Parameters.(*AllocateIpamPoolCidrInput)
	if !ok {
		return out, metadata, fmt.Errorf("expected middleware input to be of type *AllocateIpamPoolCidrInput ")
	}

	if input.ClientToken == nil {
		t, err := m.tokenProvider.GetIdempotencyToken()
		if err != nil {
			return out, metadata, err
		}
		input.ClientToken = &t
	}
	return next.HandleInitialize(ctx, in)
}
func addIdempotencyToken_opAllocateIpamPoolCidrMiddleware(stack *middleware.Stack, cfg Options) error {
	return stack.Initialize.Add(&idempotencyToken_initializeOpAllocateIpamPoolCidr{tokenProvider: cfg.IdempotencyTokenProvider}, middleware.Before)
}

func newServiceMetadataMiddleware_opAllocateIpamPoolCidr(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "AllocateIpamPoolCidr",
	}
}
