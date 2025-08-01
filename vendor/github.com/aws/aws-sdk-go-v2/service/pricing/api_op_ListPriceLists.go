// Code generated by smithy-go-codegen DO NOT EDIT.

package pricing

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

//	This feature is in preview release and is subject to change. Your use of
//
// Amazon Web Services Price List API is subject to the Beta Service Participation
// terms of the [Amazon Web Services Service Terms](Section 1.10).
//
// This returns a list of Price List references that the requester if authorized
// to view, given a ServiceCode , CurrencyCode , and an EffectiveDate . Use without
// a RegionCode filter to list Price List references from all available Amazon Web
// Services Regions. Use with a RegionCode filter to get the Price List reference
// that's specific to a specific Amazon Web Services Region. You can use the
// PriceListArn from the response to get your preferred Price List files through
// the [GetPriceListFileUrl]API.
//
// [Amazon Web Services Service Terms]: https://aws.amazon.com/service-terms/
// [GetPriceListFileUrl]: https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_pricing_GetPriceListFileUrl.html
func (c *Client) ListPriceLists(ctx context.Context, params *ListPriceListsInput, optFns ...func(*Options)) (*ListPriceListsOutput, error) {
	if params == nil {
		params = &ListPriceListsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListPriceLists", params, optFns, c.addOperationListPriceListsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListPriceListsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListPriceListsInput struct {

	// The three alphabetical character ISO-4217 currency code that the Price List
	// files are denominated in.
	//
	// This member is required.
	CurrencyCode *string

	// The date that the Price List file prices are effective from.
	//
	// This member is required.
	EffectiveDate *time.Time

	// The service code or the Savings Plan service code for the attributes that you
	// want to retrieve. For example, to get the list of applicable Amazon EC2 price
	// lists, use AmazonEC2 . For a full list of service codes containing On-Demand and
	// Reserved Instance (RI) pricing, use the [DescribeServices]API.
	//
	// To retrieve the Reserved Instance and Compute Savings Plan price lists, use
	// ComputeSavingsPlans .
	//
	// To retrieve Machine Learning Savings Plans price lists, use
	// MachineLearningSavingsPlans .
	//
	// [DescribeServices]: https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_pricing_DescribeServices.html#awscostmanagement-pricing_DescribeServices-request-FormatVersion
	//
	// This member is required.
	ServiceCode *string

	// The maximum number of results to return in the response.
	MaxResults *int32

	// The pagination token that indicates the next set of results that you want to
	// retrieve.
	NextToken *string

	// This is used to filter the Price List by Amazon Web Services Region. For
	// example, to get the price list only for the US East (N. Virginia) Region, use
	// us-east-1 . If nothing is specified, you retrieve price lists for all applicable
	// Regions. The available RegionCode list can be retrieved from [GetAttributeValues] API.
	//
	// [GetAttributeValues]: https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_pricing_GetAttributeValues.html
	RegionCode *string

	noSmithyDocumentSerde
}

type ListPriceListsOutput struct {

	// The pagination token that indicates the next set of results to retrieve.
	NextToken *string

	// The type of price list references that match your request.
	PriceLists []types.PriceList

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListPriceListsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpListPriceLists{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpListPriceLists{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "ListPriceLists"); err != nil {
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
	if err = addOpListPriceListsValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListPriceLists(options.Region), middleware.Before); err != nil {
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

// ListPriceListsPaginatorOptions is the paginator options for ListPriceLists
type ListPriceListsPaginatorOptions struct {
	// The maximum number of results to return in the response.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListPriceListsPaginator is a paginator for ListPriceLists
type ListPriceListsPaginator struct {
	options   ListPriceListsPaginatorOptions
	client    ListPriceListsAPIClient
	params    *ListPriceListsInput
	nextToken *string
	firstPage bool
}

// NewListPriceListsPaginator returns a new ListPriceListsPaginator
func NewListPriceListsPaginator(client ListPriceListsAPIClient, params *ListPriceListsInput, optFns ...func(*ListPriceListsPaginatorOptions)) *ListPriceListsPaginator {
	if params == nil {
		params = &ListPriceListsInput{}
	}

	options := ListPriceListsPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListPriceListsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListPriceListsPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListPriceLists page.
func (p *ListPriceListsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListPriceListsOutput, error) {
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
	result, err := p.client.ListPriceLists(ctx, &params, optFns...)
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

// ListPriceListsAPIClient is a client that implements the ListPriceLists
// operation.
type ListPriceListsAPIClient interface {
	ListPriceLists(context.Context, *ListPriceListsInput, ...func(*Options)) (*ListPriceListsOutput, error)
}

var _ ListPriceListsAPIClient = (*Client)(nil)

func newServiceMetadataMiddleware_opListPriceLists(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "ListPriceLists",
	}
}
