package stacks

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/vpc/orchestrator"
	utilInfra "github.com/adrianriobo/qenvs/pkg/util/infra"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

func CreateVPC(projectName, backedURL, cidr string,
	azs, publicSubnets, privateSubnets, intraSubnets []string) error {

	request := orchestrator.NetworkRequest{
		CIDR:                cidr,
		Name:                projectName,
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  publicSubnets,
		PrivateSubnetsCIDRs: privateSubnets,
		IntraSubnetsCIDRs:   intraSubnets,
		SingleNatGateway:    false}
	stack := utilInfra.Stack{
		StackName:   orchestrator.StackCreateNetworkName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault,
		DeployFunc:  request.CreateNetwork,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return err
	}
	vpcID, ok := stackResult.Outputs[orchestrator.StackCreateNetworkOutputVPCID].Value.(string)
	if !ok {
		return fmt.Errorf("error getting vpc id")
	}
	logging.Debugf("VPC has been created with id: %s", vpcID)
	return nil
}

func DestroyVPC(projectName, backedURL string) error {
	stack := utilInfra.Stack{
		StackName:   orchestrator.StackCreateNetworkName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	_, err := utilInfra.DestroyStack(stack)
	if err != nil {
		return err
	}
	logging.Debugf("VPC has been destroyed")
	return nil
}
