// Code generated by smithy-go-codegen DO NOT EDIT.

package ec2

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// Retrieves a summary of the account status report.
//
// To view the full report, download it from the Amazon S3 bucket where it was
// saved. Reports are accessible only when they have the complete status. Reports
// with other statuses ( running , cancelled , or error ) are not available in the
// S3 bucket. For more information about downloading objects from an S3 bucket, see
// [Downloading objects]in the Amazon Simple Storage Service User Guide.
//
// For more information, see [Generating the account status report for declarative policies] in the Amazon Web Services Organizations User Guide.
//
// [Downloading objects]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/download-objects.html
// [Generating the account status report for declarative policies]: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_declarative_status-report.html
func (c *Client) GetDeclarativePoliciesReportSummary(ctx context.Context, params *GetDeclarativePoliciesReportSummaryInput, optFns ...func(*Options)) (*GetDeclarativePoliciesReportSummaryOutput, error) {
	if params == nil {
		params = &GetDeclarativePoliciesReportSummaryInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetDeclarativePoliciesReportSummary", params, optFns, c.addOperationGetDeclarativePoliciesReportSummaryMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetDeclarativePoliciesReportSummaryOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetDeclarativePoliciesReportSummaryInput struct {

	// The ID of the report.
	//
	// This member is required.
	ReportId *string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	noSmithyDocumentSerde
}

type GetDeclarativePoliciesReportSummaryOutput struct {

	// The attributes described in the report.
	AttributeSummaries []types.AttributeSummary

	// The time when the report generation ended.
	EndTime *time.Time

	// The total number of accounts associated with the specified targetId .
	NumberOfAccounts *int32

	// The number of accounts where attributes could not be retrieved in any Region.
	NumberOfFailedAccounts *int32

	// The ID of the report.
	ReportId *string

	// The name of the Amazon S3 bucket where the report is located.
	S3Bucket *string

	// The prefix for your S3 object.
	S3Prefix *string

	// The time when the report generation started.
	StartTime *time.Time

	// The root ID, organizational unit ID, or account ID.
	//
	// Format:
	//
	//   - For root: r-ab12
	//
	//   - For OU: ou-ab12-cdef1234
	//
	//   - For account: 123456789012
	TargetId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetDeclarativePoliciesReportSummaryMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsEc2query_serializeOpGetDeclarativePoliciesReportSummary{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpGetDeclarativePoliciesReportSummary{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GetDeclarativePoliciesReportSummary"); err != nil {
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
	if err = addOpGetDeclarativePoliciesReportSummaryValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetDeclarativePoliciesReportSummary(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opGetDeclarativePoliciesReportSummary(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "GetDeclarativePoliciesReportSummary",
	}
}
