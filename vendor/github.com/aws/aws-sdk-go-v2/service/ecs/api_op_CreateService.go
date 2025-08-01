// Code generated by smithy-go-codegen DO NOT EDIT.

package ecs

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Runs and maintains your desired number of tasks from a specified task
// definition. If the number of tasks running in a service drops below the
// desiredCount , Amazon ECS runs another copy of the task in the specified
// cluster. To update an existing service, use [UpdateService].
//
// On March 21, 2024, a change was made to resolve the task definition revision
// before authorization. When a task definition revision is not specified,
// authorization will occur using the latest revision of a task definition.
//
// Amazon Elastic Inference (EI) is no longer available to customers.
//
// In addition to maintaining the desired count of tasks in your service, you can
// optionally run your service behind one or more load balancers. The load
// balancers distribute traffic across the tasks that are associated with the
// service. For more information, see [Service load balancing]in the Amazon Elastic Container Service
// Developer Guide.
//
// You can attach Amazon EBS volumes to Amazon ECS tasks by configuring the volume
// when creating or updating a service. volumeConfigurations is only supported for
// REPLICA service and not DAEMON service. For more information, see [Amazon EBS volumes]in the Amazon
// Elastic Container Service Developer Guide.
//
// Tasks for services that don't use a load balancer are considered healthy if
// they're in the RUNNING state. Tasks for services that use a load balancer are
// considered healthy if they're in the RUNNING state and are reported as healthy
// by the load balancer.
//
// There are two service scheduler strategies available:
//
//   - REPLICA - The replica scheduling strategy places and maintains your desired
//     number of tasks across your cluster. By default, the service scheduler spreads
//     tasks across Availability Zones. You can use task placement strategies and
//     constraints to customize task placement decisions. For more information, see [Service scheduler concepts]
//     in the Amazon Elastic Container Service Developer Guide.
//
//   - DAEMON - The daemon scheduling strategy deploys exactly one task on each
//     active container instance that meets all of the task placement constraints that
//     you specify in your cluster. The service scheduler also evaluates the task
//     placement constraints for running tasks. It also stops tasks that don't meet the
//     placement constraints. When using this strategy, you don't need to specify a
//     desired number of tasks, a task placement strategy, or use Service Auto Scaling
//     policies. For more information, see [Amazon ECS services]in the Amazon Elastic Container Service
//     Developer Guide.
//
// The deployment controller is the mechanism that determines how tasks are
// deployed for your service. The valid options are:
//
//   - ECS
//
// When you create a service which uses the ECS deployment controller, you can
//
//	choose between the following deployment strategies (which you can set in the “
//	strategy ” field in “ deploymentConfiguration ”): :
//
//	- ROLLING : When you create a service which uses the rolling update ( ROLLING
//	) deployment strategy, the Amazon ECS service scheduler replaces the currently
//	running tasks with new tasks. The number of tasks that Amazon ECS adds or
//	removes from the service during a rolling update is controlled by the service
//	deployment configuration. For more information, see [Deploy Amazon ECS services by replacing tasks]in the Amazon Elastic
//	Container Service Developer Guide.
//
// Rolling update deployments are best suited for the following scenarios:
//
//   - Gradual service updates: You need to update your service incrementally
//     without taking the entire service offline at once.
//
//   - Limited resource requirements: You want to avoid the additional resource
//     costs of running two complete environments simultaneously (as required by
//     blue/green deployments).
//
//   - Acceptable deployment time: Your application can tolerate a longer
//     deployment process, as rolling updates replace tasks one by one.
//
//   - No need for instant roll back: Your service can tolerate a rollback process
//     that takes minutes rather than seconds.
//
//   - Simple deployment process: You prefer a straightforward deployment approach
//     without the complexity of managing multiple environments, target groups, and
//     listeners.
//
//   - No load balancer requirement: Your service doesn't use or require a load
//     balancer, Application Load Balancer, Network Load Balancer, or Service Connect
//     (which are required for blue/green deployments).
//
//   - Stateful applications: Your application maintains state that makes it
//     difficult to run two parallel environments.
//
//   - Cost sensitivity: You want to minimize deployment costs by not running
//     duplicate environments during deployment.
//
// Rolling updates are the default deployment strategy for services and provide a
//
//	balance between deployment safety and resource efficiency for many common
//	application scenarios.
//
//	- BLUE_GREEN : A blue/green deployment strategy ( BLUE_GREEN ) is a release
//	methodology that reduces downtime and risk by running two identical production
//	environments called blue and green. With Amazon ECS blue/green deployments, you
//	can validate new service revisions before directing production traffic to them.
//	This approach provides a safer way to deploy changes with the ability to quickly
//	roll back if needed. For more information, see [Amazon ECS blue/green deployments]in the Amazon Elastic
//	Container Service Developer Guide.
//
// Amazon ECS blue/green deployments are best suited for the following scenarios:
//
//   - Service validation: When you need to validate new service revisions before
//     directing production traffic to them
//
//   - Zero downtime: When your service requires zero-downtime deployments
//
//   - Instant roll back: When you need the ability to quickly roll back if issues
//     are detected
//
//   - Load balancer requirement: When your service uses Application Load
//     Balancer, Network Load Balancer, or Service Connect
//
//   - External
//
// Use a third-party deployment controller.
//
//   - Blue/green deployment (powered by CodeDeploy)
//
// CodeDeploy installs an updated version of the application as a new replacement
//
//	task set and reroutes production traffic from the original application task set
//	to the replacement task set. The original task set is terminated after a
//	successful deployment. Use this deployment controller to verify a new deployment
//	of a service before sending production traffic to it.
//
// When creating a service that uses the EXTERNAL deployment controller, you can
// specify only parameters that aren't controlled at the task set level. The only
// required parameter is the service name. You control your services using the [CreateTaskSet].
// For more information, see [Amazon ECS deployment types]in the Amazon Elastic Container Service Developer
// Guide.
//
// When the service scheduler launches new tasks, it determines task placement.
// For information about task placement and task placement strategies, see [Amazon ECS task placement]in the
// Amazon Elastic Container Service Developer Guide
//
// [Amazon ECS task placement]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-placement.html
// [Service scheduler concepts]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html
// [Amazon ECS deployment types]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-types.html
// [UpdateService]: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html
// [CreateTaskSet]: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_CreateTaskSet.html
// [Amazon ECS services]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html
// [Service load balancing]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-load-balancing.html
// [Amazon EBS volumes]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ebs-volumes.html#ebs-volume-types
//
// [Amazon ECS blue/green deployments]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-type-blue-green.html
// [Deploy Amazon ECS services by replacing tasks]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-type-ecs.html
func (c *Client) CreateService(ctx context.Context, params *CreateServiceInput, optFns ...func(*Options)) (*CreateServiceOutput, error) {
	if params == nil {
		params = &CreateServiceInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "CreateService", params, optFns, c.addOperationCreateServiceMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*CreateServiceOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type CreateServiceInput struct {

	// The name of your service. Up to 255 letters (uppercase and lowercase), numbers,
	// underscores, and hyphens are allowed. Service names must be unique within a
	// cluster, but you can have similarly named services in multiple clusters within a
	// Region or across multiple Regions.
	//
	// This member is required.
	ServiceName *string

	// Indicates whether to use Availability Zone rebalancing for the service.
	//
	// For more information, see [Balancing an Amazon ECS service across Availability Zones] in the Amazon Elastic Container Service Developer
	// Guide .
	//
	// [Balancing an Amazon ECS service across Availability Zones]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-rebalancing.html
	AvailabilityZoneRebalancing types.AvailabilityZoneRebalancing

	// The capacity provider strategy to use for the service.
	//
	// If a capacityProviderStrategy is specified, the launchType parameter must be
	// omitted. If no capacityProviderStrategy or launchType is specified, the
	// defaultCapacityProviderStrategy for the cluster is used.
	//
	// A capacity provider strategy can contain a maximum of 20 capacity providers.
	CapacityProviderStrategy []types.CapacityProviderStrategyItem

	// An identifier that you provide to ensure the idempotency of the request. It
	// must be unique and is case sensitive. Up to 36 ASCII characters in the range of
	// 33-126 (inclusive) are allowed.
	ClientToken *string

	// The short name or full Amazon Resource Name (ARN) of the cluster that you run
	// your service on. If you do not specify a cluster, the default cluster is
	// assumed.
	Cluster *string

	// Optional deployment parameters that control how many tasks run during the
	// deployment and the ordering of stopping and starting tasks.
	DeploymentConfiguration *types.DeploymentConfiguration

	// The deployment controller to use for the service. If no deployment controller
	// is specified, the default value of ECS is used.
	DeploymentController *types.DeploymentController

	// The number of instantiations of the specified task definition to place and keep
	// running in your service.
	//
	// This is required if schedulingStrategy is REPLICA or isn't specified. If
	// schedulingStrategy is DAEMON then this isn't required.
	DesiredCount *int32

	// Specifies whether to turn on Amazon ECS managed tags for the tasks within the
	// service. For more information, see [Tagging your Amazon ECS resources]in the Amazon Elastic Container Service
	// Developer Guide.
	//
	// When you use Amazon ECS managed tags, you must set the propagateTags request
	// parameter.
	//
	// [Tagging your Amazon ECS resources]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-using-tags.html
	EnableECSManagedTags bool

	// Determines whether the execute command functionality is turned on for the
	// service. If true , this enables execute command functionality on all containers
	// in the service tasks.
	EnableExecuteCommand bool

	// The period of time, in seconds, that the Amazon ECS service scheduler ignores
	// unhealthy Elastic Load Balancing, VPC Lattice, and container health checks after
	// a task has first started. If you don't specify a health check grace period
	// value, the default value of 0 is used. If you don't use any of the health
	// checks, then healthCheckGracePeriodSeconds is unused.
	//
	// If your service's tasks take a while to start and respond to health checks, you
	// can specify a health check grace period of up to 2,147,483,647 seconds (about 69
	// years). During that time, the Amazon ECS service scheduler ignores health check
	// status. This grace period can prevent the service scheduler from marking tasks
	// as unhealthy and stopping them before they have time to come up.
	HealthCheckGracePeriodSeconds *int32

	// The infrastructure that you run your service on. For more information, see [Amazon ECS launch types] in
	// the Amazon Elastic Container Service Developer Guide.
	//
	// The FARGATE launch type runs your tasks on Fargate On-Demand infrastructure.
	//
	// Fargate Spot infrastructure is available for use but a capacity provider
	// strategy must be used. For more information, see [Fargate capacity providers]in the Amazon ECS Developer
	// Guide.
	//
	// The EC2 launch type runs your tasks on Amazon EC2 instances registered to your
	// cluster.
	//
	// The EXTERNAL launch type runs your tasks on your on-premises server or virtual
	// machine (VM) capacity registered to your cluster.
	//
	// A service can use either a launch type or a capacity provider strategy. If a
	// launchType is specified, the capacityProviderStrategy parameter must be omitted.
	//
	// [Amazon ECS launch types]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/launch_types.html
	// [Fargate capacity providers]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/fargate-capacity-providers.html
	LaunchType types.LaunchType

	// A load balancer object representing the load balancers to use with your
	// service. For more information, see [Service load balancing]in the Amazon Elastic Container Service
	// Developer Guide.
	//
	// If the service uses the rolling update ( ECS ) deployment controller and using
	// either an Application Load Balancer or Network Load Balancer, you must specify
	// one or more target group ARNs to attach to the service. The service-linked role
	// is required for services that use multiple target groups. For more information,
	// see [Using service-linked roles for Amazon ECS]in the Amazon Elastic Container Service Developer Guide.
	//
	// If the service uses the CODE_DEPLOY deployment controller, the service is
	// required to use either an Application Load Balancer or Network Load Balancer.
	// When creating an CodeDeploy deployment group, you specify two target groups
	// (referred to as a targetGroupPair ). During a deployment, CodeDeploy determines
	// which task set in your service has the status PRIMARY , and it associates one
	// target group with it. Then, it also associates the other target group with the
	// replacement task set. The load balancer can also have up to two listeners: a
	// required listener for production traffic and an optional listener that you can
	// use to perform validation tests with Lambda functions before routing production
	// traffic to it.
	//
	// If you use the CODE_DEPLOY deployment controller, these values can be changed
	// when updating the service.
	//
	// For Application Load Balancers and Network Load Balancers, this object must
	// contain the load balancer target group ARN, the container name, and the
	// container port to access from the load balancer. The container name must be as
	// it appears in a container definition. The load balancer name parameter must be
	// omitted. When a task from this service is placed on a container instance, the
	// container instance and port combination is registered as a target in the target
	// group that's specified here.
	//
	// For Classic Load Balancers, this object must contain the load balancer name,
	// the container name , and the container port to access from the load balancer.
	// The container name must be as it appears in a container definition. The target
	// group ARN parameter must be omitted. When a task from this service is placed on
	// a container instance, the container instance is registered with the load
	// balancer that's specified here.
	//
	// Services with tasks that use the awsvpc network mode (for example, those with
	// the Fargate launch type) only support Application Load Balancers and Network
	// Load Balancers. Classic Load Balancers aren't supported. Also, when you create
	// any target groups for these services, you must choose ip as the target type,
	// not instance . This is because tasks that use the awsvpc network mode are
	// associated with an elastic network interface, not an Amazon EC2 instance.
	//
	// [Service load balancing]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-load-balancing.html
	// [Using service-linked roles for Amazon ECS]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/using-service-linked-roles.html
	LoadBalancers []types.LoadBalancer

	// The network configuration for the service. This parameter is required for task
	// definitions that use the awsvpc network mode to receive their own elastic
	// network interface, and it isn't supported for other network modes. For more
	// information, see [Task networking]in the Amazon Elastic Container Service Developer Guide.
	//
	// [Task networking]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html
	NetworkConfiguration *types.NetworkConfiguration

	// An array of placement constraint objects to use for tasks in your service. You
	// can specify a maximum of 10 constraints for each task. This limit includes
	// constraints in the task definition and those specified at runtime.
	PlacementConstraints []types.PlacementConstraint

	// The placement strategy objects to use for tasks in your service. You can
	// specify a maximum of 5 strategy rules for each service.
	PlacementStrategy []types.PlacementStrategy

	// The platform version that your tasks in the service are running on. A platform
	// version is specified only for tasks using the Fargate launch type. If one isn't
	// specified, the LATEST platform version is used. For more information, see [Fargate platform versions] in
	// the Amazon Elastic Container Service Developer Guide.
	//
	// [Fargate platform versions]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/platform_versions.html
	PlatformVersion *string

	// Specifies whether to propagate the tags from the task definition to the task.
	// If no value is specified, the tags aren't propagated. Tags can only be
	// propagated to the task during task creation. To add tags to a task after task
	// creation, use the [TagResource]API action.
	//
	// You must set this to a value other than NONE when you use Cost Explorer. For
	// more information, see [Amazon ECS usage reports]in the Amazon Elastic Container Service Developer Guide.
	//
	// The default is NONE .
	//
	// [TagResource]: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_TagResource.html
	// [Amazon ECS usage reports]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/usage-reports.html
	PropagateTags types.PropagateTags

	// The name or full Amazon Resource Name (ARN) of the IAM role that allows Amazon
	// ECS to make calls to your load balancer on your behalf. This parameter is only
	// permitted if you are using a load balancer with your service and your task
	// definition doesn't use the awsvpc network mode. If you specify the role
	// parameter, you must also specify a load balancer object with the loadBalancers
	// parameter.
	//
	// If your account has already created the Amazon ECS service-linked role, that
	// role is used for your service unless you specify a role here. The service-linked
	// role is required if your task definition uses the awsvpc network mode or if the
	// service is configured to use service discovery, an external deployment
	// controller, multiple target groups, or Elastic Inference accelerators in which
	// case you don't specify a role here. For more information, see [Using service-linked roles for Amazon ECS]in the Amazon
	// Elastic Container Service Developer Guide.
	//
	// If your specified role has a path other than / , then you must either specify
	// the full role ARN (this is recommended) or prefix the role name with the path.
	// For example, if a role with the name bar has a path of /foo/ then you would
	// specify /foo/bar as the role name. For more information, see [Friendly names and paths] in the IAM User
	// Guide.
	//
	// [Friendly names and paths]: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-friendly-names
	// [Using service-linked roles for Amazon ECS]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/using-service-linked-roles.html
	Role *string

	// The scheduling strategy to use for the service. For more information, see [Services].
	//
	// There are two service scheduler strategies available:
	//
	//   - REPLICA -The replica scheduling strategy places and maintains the desired
	//   number of tasks across your cluster. By default, the service scheduler spreads
	//   tasks across Availability Zones. You can use task placement strategies and
	//   constraints to customize task placement decisions. This scheduler strategy is
	//   required if the service uses the CODE_DEPLOY or EXTERNAL deployment controller
	//   types.
	//
	//   - DAEMON -The daemon scheduling strategy deploys exactly one task on each
	//   active container instance that meets all of the task placement constraints that
	//   you specify in your cluster. The service scheduler also evaluates the task
	//   placement constraints for running tasks and will stop tasks that don't meet the
	//   placement constraints. When you're using this strategy, you don't need to
	//   specify a desired number of tasks, a task placement strategy, or use Service
	//   Auto Scaling policies.
	//
	// Tasks using the Fargate launch type or the CODE_DEPLOY or EXTERNAL deployment
	//   controller types don't support the DAEMON scheduling strategy.
	//
	// [Services]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html
	SchedulingStrategy types.SchedulingStrategy

	// The configuration for this service to discover and connect to services, and be
	// discovered by, and connected from, other services within a namespace.
	//
	// Tasks that run in a namespace can use short names to connect to services in the
	// namespace. Tasks can connect to services across all of the clusters in the
	// namespace. Tasks connect through a managed proxy container that collects logs
	// and metrics for increased visibility. Only the tasks that Amazon ECS services
	// create are supported with Service Connect. For more information, see [Service Connect]in the
	// Amazon Elastic Container Service Developer Guide.
	//
	// [Service Connect]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-connect.html
	ServiceConnectConfiguration *types.ServiceConnectConfiguration

	// The details of the service discovery registry to associate with this service.
	// For more information, see [Service discovery].
	//
	// Each service may be associated with one service registry. Multiple service
	// registries for each service isn't supported.
	//
	// [Service discovery]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-discovery.html
	ServiceRegistries []types.ServiceRegistry

	// The metadata that you apply to the service to help you categorize and organize
	// them. Each tag consists of a key and an optional value, both of which you
	// define. When a service is deleted, the tags are deleted as well.
	//
	// The following basic restrictions apply to tags:
	//
	//   - Maximum number of tags per resource - 50
	//
	//   - For each resource, each tag key must be unique, and each tag key can have
	//   only one value.
	//
	//   - Maximum key length - 128 Unicode characters in UTF-8
	//
	//   - Maximum value length - 256 Unicode characters in UTF-8
	//
	//   - If your tagging schema is used across multiple services and resources,
	//   remember that other services may have restrictions on allowed characters.
	//   Generally allowed characters are: letters, numbers, and spaces representable in
	//   UTF-8, and the following characters: + - = . _ : / @.
	//
	//   - Tag keys and values are case-sensitive.
	//
	//   - Do not use aws: , AWS: , or any upper or lowercase combination of such as a
	//   prefix for either keys or values as it is reserved for Amazon Web Services use.
	//   You cannot edit or delete tag keys or values with this prefix. Tags with this
	//   prefix do not count against your tags per resource limit.
	Tags []types.Tag

	// The family and revision ( family:revision ) or full ARN of the task definition
	// to run in your service. If a revision isn't specified, the latest ACTIVE
	// revision is used.
	//
	// A task definition must be specified if the service uses either the ECS or
	// CODE_DEPLOY deployment controllers.
	//
	// For more information about deployment types, see [Amazon ECS deployment types].
	//
	// [Amazon ECS deployment types]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-types.html
	TaskDefinition *string

	// The configuration for a volume specified in the task definition as a volume
	// that is configured at launch time. Currently, the only supported volume type is
	// an Amazon EBS volume.
	VolumeConfigurations []types.ServiceVolumeConfiguration

	// The VPC Lattice configuration for the service being created.
	VpcLatticeConfigurations []types.VpcLatticeConfiguration

	noSmithyDocumentSerde
}

type CreateServiceOutput struct {

	// The full description of your service following the create call.
	//
	// A service will return either a capacityProviderStrategy or launchType
	// parameter, but not both, depending where one was specified when it was created.
	//
	// If a service is using the ECS deployment controller, the deploymentController
	// and taskSets parameters will not be returned.
	//
	// if the service uses the CODE_DEPLOY deployment controller, the
	// deploymentController , taskSets and deployments parameters will be returned,
	// however the deployments parameter will be an empty list.
	Service *types.Service

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationCreateServiceMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpCreateService{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpCreateService{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "CreateService"); err != nil {
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
	if err = addOpCreateServiceValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opCreateService(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opCreateService(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "CreateService",
	}
}
