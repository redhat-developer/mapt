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

// Displays details about an import virtual machine or import snapshot tasks that
// are already created.
func (c *Client) DescribeImportImageTasks(ctx context.Context, params *DescribeImportImageTasksInput, optFns ...func(*Options)) (*DescribeImportImageTasksOutput, error) {
	if params == nil {
		params = &DescribeImportImageTasksInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DescribeImportImageTasks", params, optFns, c.addOperationDescribeImportImageTasksMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DescribeImportImageTasksOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DescribeImportImageTasksInput struct {

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// Filter tasks using the task-state filter and one of the following values: active
	// , completed , deleting , or deleted .
	Filters []types.Filter

	// The IDs of the import image tasks.
	ImportTaskIds []string

	// The maximum number of results to return in a single call.
	MaxResults *int32

	// A token that indicates the next page of results.
	NextToken *string

	noSmithyDocumentSerde
}

type DescribeImportImageTasksOutput struct {

	// A list of zero or more import image tasks that are currently active or were
	// completed or canceled in the previous 7 days.
	ImportImageTasks []types.ImportImageTask

	// The token to use to get the next page of results. This value is null when there
	// are no more results to return.
	NextToken *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDescribeImportImageTasksMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsEc2query_serializeOpDescribeImportImageTasks{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpDescribeImportImageTasks{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "DescribeImportImageTasks"); err != nil {
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDescribeImportImageTasks(options.Region), middleware.Before); err != nil {
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

// DescribeImportImageTasksPaginatorOptions is the paginator options for
// DescribeImportImageTasks
type DescribeImportImageTasksPaginatorOptions struct {
	// The maximum number of results to return in a single call.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// DescribeImportImageTasksPaginator is a paginator for DescribeImportImageTasks
type DescribeImportImageTasksPaginator struct {
	options   DescribeImportImageTasksPaginatorOptions
	client    DescribeImportImageTasksAPIClient
	params    *DescribeImportImageTasksInput
	nextToken *string
	firstPage bool
}

// NewDescribeImportImageTasksPaginator returns a new
// DescribeImportImageTasksPaginator
func NewDescribeImportImageTasksPaginator(client DescribeImportImageTasksAPIClient, params *DescribeImportImageTasksInput, optFns ...func(*DescribeImportImageTasksPaginatorOptions)) *DescribeImportImageTasksPaginator {
	if params == nil {
		params = &DescribeImportImageTasksInput{}
	}

	options := DescribeImportImageTasksPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &DescribeImportImageTasksPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *DescribeImportImageTasksPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next DescribeImportImageTasks page.
func (p *DescribeImportImageTasksPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*DescribeImportImageTasksOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxResults = limit

	optFns = append([]func(*Options){
		addIsPaginatorUserAgent,
	}, optFns...)
	result, err := p.client.DescribeImportImageTasks(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

// DescribeImportImageTasksAPIClient is a client that implements the
// DescribeImportImageTasks operation.
type DescribeImportImageTasksAPIClient interface {
	DescribeImportImageTasks(context.Context, *DescribeImportImageTasksInput, ...func(*Options)) (*DescribeImportImageTasksOutput, error)
}

var _ DescribeImportImageTasksAPIClient = (*Client)(nil)

func newServiceMetadataMiddleware_opDescribeImportImageTasks(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "DescribeImportImageTasks",
	}
}
