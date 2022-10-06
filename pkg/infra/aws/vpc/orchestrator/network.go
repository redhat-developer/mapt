package orchestrator

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/vpc/stacks"
	utilInfra "github.com/adrianriobo/qenvs/pkg/util/infra"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

func CreateNetwork(projectName, backedURL, cidr string,
	azs, publicSubnets, privateSubnets, intraSubnets []string) error {

	request := stacks.NetworkRequest{
		CIDR:                cidr,
		Name:                projectName,
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  publicSubnets,
		PrivateSubnetsCIDRs: privateSubnets,
		IntraSubnetsCIDRs:   intraSubnets,
		SingleNatGateway:    false}
	stack := utilInfra.Stack{
		StackName:   stacks.StackCreateNetworkName,
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
	vpcID, ok := stackResult.Outputs[stacks.StackCreateNetworkOutputVPCID].Value.(string)
	if !ok {
		return fmt.Errorf("error getting vpc id")
	}
	logging.Debugf("VPC has been created with id: %s", vpcID)
	return nil
}

func DestroyNetwork(projectName, backedURL string) error {
	stack := utilInfra.Stack{
		StackName:   stacks.StackCreateNetworkName,
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
