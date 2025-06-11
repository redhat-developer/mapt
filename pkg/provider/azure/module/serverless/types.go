package serverless

import "github.com/pulumi/pulumi-azure-native-sdk/resources/v3"

// https://learn.microsoft.com/en-us/rest/api/resource-manager/containerapps/jobs/create-or-update?view=rest-resource-manager-containerapps-2025-01-01&tabs=HTTP#containerresources
const (
	LimitCPU    = "1" // azure expects it as 2.0
	LimitMemory = "2.0Gi"

	maptServerlessDefaultPrefix = "mapt-serverless-manager"
)

type serverlessRequestArgs struct {
	region             string
	command            string
	scheduleExpression string
	resourceGroup      *resources.ResourceGroup

	// optional params in case we create serverless inside a stack
	prefix, componentID string
}
