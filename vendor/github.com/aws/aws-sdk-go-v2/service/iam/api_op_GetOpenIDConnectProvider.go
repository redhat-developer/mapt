// Code generated by smithy-go-codegen DO NOT EDIT.

package iam

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// Returns information about the specified OpenID Connect (OIDC) provider resource
// object in IAM.
func (c *Client) GetOpenIDConnectProvider(ctx context.Context, params *GetOpenIDConnectProviderInput, optFns ...func(*Options)) (*GetOpenIDConnectProviderOutput, error) {
	if params == nil {
		params = &GetOpenIDConnectProviderInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetOpenIDConnectProvider", params, optFns, c.addOperationGetOpenIDConnectProviderMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetOpenIDConnectProviderOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetOpenIDConnectProviderInput struct {

	// The Amazon Resource Name (ARN) of the OIDC provider resource object in IAM to
	// get information for. You can get a list of OIDC provider resource ARNs by using
	// the [ListOpenIDConnectProviders]operation.
	//
	// For more information about ARNs, see [Amazon Resource Names (ARNs)] in the Amazon Web Services General
	// Reference.
	//
	// [ListOpenIDConnectProviders]: https://docs.aws.amazon.com/IAM/latest/APIReference/API_ListOpenIDConnectProviders.html
	// [Amazon Resource Names (ARNs)]: https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
	//
	// This member is required.
	OpenIDConnectProviderArn *string

	noSmithyDocumentSerde
}

// Contains the response to a successful [GetOpenIDConnectProvider] request.
//
// [GetOpenIDConnectProvider]: https://docs.aws.amazon.com/IAM/latest/APIReference/API_GetOpenIDConnectProvider.html
type GetOpenIDConnectProviderOutput struct {

	// A list of client IDs (also known as audiences) that are associated with the
	// specified IAM OIDC provider resource object. For more information, see [CreateOpenIDConnectProvider].
	//
	// [CreateOpenIDConnectProvider]: https://docs.aws.amazon.com/IAM/latest/APIReference/API_CreateOpenIDConnectProvider.html
	ClientIDList []string

	// The date and time when the IAM OIDC provider resource object was created in the
	// Amazon Web Services account.
	CreateDate *time.Time

	// A list of tags that are attached to the specified IAM OIDC provider. The
	// returned list of tags is sorted by tag key. For more information about tagging,
	// see [Tagging IAM resources]in the IAM User Guide.
	//
	// [Tagging IAM resources]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html
	Tags []types.Tag

	// A list of certificate thumbprints that are associated with the specified IAM
	// OIDC provider resource object. For more information, see [CreateOpenIDConnectProvider].
	//
	// [CreateOpenIDConnectProvider]: https://docs.aws.amazon.com/IAM/latest/APIReference/API_CreateOpenIDConnectProvider.html
	ThumbprintList []string

	// The URL that the IAM OIDC provider resource object is associated with. For more
	// information, see [CreateOpenIDConnectProvider].
	//
	// [CreateOpenIDConnectProvider]: https://docs.aws.amazon.com/IAM/latest/APIReference/API_CreateOpenIDConnectProvider.html
	Url *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetOpenIDConnectProviderMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsquery_serializeOpGetOpenIDConnectProvider{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsquery_deserializeOpGetOpenIDConnectProvider{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GetOpenIDConnectProvider"); err != nil {
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
	if err = addOpGetOpenIDConnectProviderValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetOpenIDConnectProvider(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opGetOpenIDConnectProvider(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "GetOpenIDConnectProvider",
	}
}
