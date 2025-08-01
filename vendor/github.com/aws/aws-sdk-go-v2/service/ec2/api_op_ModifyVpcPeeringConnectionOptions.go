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

// Modifies the VPC peering connection options on one side of a VPC peering
// connection.
//
// If the peered VPCs are in the same Amazon Web Services account, you can enable
// DNS resolution for queries from the local VPC. This ensures that queries from
// the local VPC resolve to private IP addresses in the peer VPC. This option is
// not available if the peered VPCs are in different Amazon Web Services accounts
// or different Regions. For peered VPCs in different Amazon Web Services accounts,
// each Amazon Web Services account owner must initiate a separate request to
// modify the peering connection options. For inter-region peering connections, you
// must use the Region for the requester VPC to modify the requester VPC peering
// options and the Region for the accepter VPC to modify the accepter VPC peering
// options. To verify which VPCs are the accepter and the requester for a VPC
// peering connection, use the DescribeVpcPeeringConnectionscommand.
func (c *Client) ModifyVpcPeeringConnectionOptions(ctx context.Context, params *ModifyVpcPeeringConnectionOptionsInput, optFns ...func(*Options)) (*ModifyVpcPeeringConnectionOptionsOutput, error) {
	if params == nil {
		params = &ModifyVpcPeeringConnectionOptionsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ModifyVpcPeeringConnectionOptions", params, optFns, c.addOperationModifyVpcPeeringConnectionOptionsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ModifyVpcPeeringConnectionOptionsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ModifyVpcPeeringConnectionOptionsInput struct {

	// The ID of the VPC peering connection.
	//
	// This member is required.
	VpcPeeringConnectionId *string

	// The VPC peering connection options for the accepter VPC.
	AccepterPeeringConnectionOptions *types.PeeringConnectionOptionsRequest

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// The VPC peering connection options for the requester VPC.
	RequesterPeeringConnectionOptions *types.PeeringConnectionOptionsRequest

	noSmithyDocumentSerde
}

type ModifyVpcPeeringConnectionOptionsOutput struct {

	// Information about the VPC peering connection options for the accepter VPC.
	AccepterPeeringConnectionOptions *types.PeeringConnectionOptions

	// Information about the VPC peering connection options for the requester VPC.
	RequesterPeeringConnectionOptions *types.PeeringConnectionOptions

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationModifyVpcPeeringConnectionOptionsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsEc2query_serializeOpModifyVpcPeeringConnectionOptions{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpModifyVpcPeeringConnectionOptions{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "ModifyVpcPeeringConnectionOptions"); err != nil {
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
	if err = addOpModifyVpcPeeringConnectionOptionsValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opModifyVpcPeeringConnectionOptions(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opModifyVpcPeeringConnectionOptions(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "ModifyVpcPeeringConnectionOptions",
	}
}
