package iam

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

const (
	stackName = "iam-manager"

	outputAccessKey = "accessKey"
	outputSecretKey = "secretKey"
)

type iamRequestArgs struct {
	// need this to find the right ECS cluster to run this serverless
	name string
	// command and scheduling to be used for it
	policyContent *string
	// optional params in case we create serverless inside a stack
	prefix, componentID string
	// Dependecies
	dependsOn []pulumi.Resource
}
