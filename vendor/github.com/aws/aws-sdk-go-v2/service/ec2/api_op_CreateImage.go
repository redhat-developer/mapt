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

// Creates an Amazon EBS-backed AMI from an Amazon EBS-backed instance that is
// either running or stopped.
//
// If you customized your instance with instance store volumes or Amazon EBS
// volumes in addition to the root device volume, the new AMI contains block device
// mapping information for those volumes. When you launch an instance from this new
// AMI, the instance automatically launches with those additional volumes.
//
// The location of the source instance determines where you can create the
// snapshots of the AMI:
//
//   - If the source instance is in a Region, you must create the snapshots in the
//     same Region as the instance.
//
//   - If the source instance is in a Local Zone, you can create the snapshots in
//     the same Local Zone or in its parent Region.
//
// For more information, see [Create an Amazon EBS-backed AMI] in the Amazon Elastic Compute Cloud User Guide.
//
// [Create an Amazon EBS-backed AMI]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/creating-an-ami-ebs.html
func (c *Client) CreateImage(ctx context.Context, params *CreateImageInput, optFns ...func(*Options)) (*CreateImageOutput, error) {
	if params == nil {
		params = &CreateImageInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "CreateImage", params, optFns, c.addOperationCreateImageMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*CreateImageOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type CreateImageInput struct {

	// The ID of the instance.
	//
	// This member is required.
	InstanceId *string

	// A name for the new image.
	//
	// Constraints: 3-128 alphanumeric characters, parentheses (()), square brackets
	// ([]), spaces ( ), periods (.), slashes (/), dashes (-), single quotes ('),
	// at-signs (@), or underscores(_)
	//
	// This member is required.
	Name *string

	// The block device mappings.
	//
	// When using the CreateImage action:
	//
	//   - You can't change the volume size using the VolumeSize parameter. If you
	//   want a different volume size, you must first change the volume size of the
	//   source instance.
	//
	//   - You can't modify the encryption status of existing volumes or snapshots. To
	//   create an AMI with volumes or snapshots that have a different encryption status
	//   (for example, where the source volume and snapshots are unencrypted, and you
	//   want to create an AMI with encrypted volumes or snapshots), copy the image
	//   instead.
	//
	//   - The only option that can be changed for existing mappings or snapshots is
	//   DeleteOnTermination .
	BlockDeviceMappings []types.BlockDeviceMapping

	// A description for the new image.
	Description *string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// Indicates whether or not the instance should be automatically rebooted before
	// creating the image. Specify one of the following values:
	//
	//   - true - The instance is not rebooted before creating the image. This creates
	//   crash-consistent snapshots that include only the data that has been written to
	//   the volumes at the time the snapshots are created. Buffered data and data in
	//   memory that has not yet been written to the volumes is not included in the
	//   snapshots.
	//
	//   - false - The instance is rebooted before creating the image. This ensures
	//   that all buffered data and data in memory is written to the volumes before the
	//   snapshots are created.
	//
	// Default: false
	NoReboot *bool

	// Only supported for instances in Local Zones. If the source instance is not in a
	// Local Zone, omit this parameter.
	//
	// The Amazon S3 location where the snapshots will be stored.
	//
	//   - To create local snapshots in the same Local Zone as the source instance,
	//   specify local .
	//
	//   - To create regional snapshots in the parent Region of the Local Zone,
	//   specify regional or omit this parameter.
	//
	// Default: regional
	SnapshotLocation types.SnapshotLocationEnum

	// The tags to apply to the AMI and snapshots on creation. You can tag the AMI,
	// the snapshots, or both.
	//
	//   - To tag the AMI, the value for ResourceType must be image .
	//
	//   - To tag the snapshots that are created of the root volume and of other
	//   Amazon EBS volumes that are attached to the instance, the value for
	//   ResourceType must be snapshot . The same tag is applied to all of the
	//   snapshots that are created.
	//
	// If you specify other values for ResourceType , the request fails.
	//
	// To tag an AMI or snapshot after it has been created, see [CreateTags].
	//
	// [CreateTags]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateTags.html
	TagSpecifications []types.TagSpecification

	noSmithyDocumentSerde
}

type CreateImageOutput struct {

	// The ID of the new AMI.
	ImageId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationCreateImageMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsEc2query_serializeOpCreateImage{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpCreateImage{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "CreateImage"); err != nil {
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
	if err = addOpCreateImageValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opCreateImage(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opCreateImage(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "CreateImage",
	}
}
