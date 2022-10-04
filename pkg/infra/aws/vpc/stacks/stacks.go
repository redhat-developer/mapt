package stacks

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/vpc/orchestrator"
	infraUtil "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

func CreateVPC(projectName, backedURL, cidr string,
	azs, privateSubnets, publicSubnets []string) error {

	request := orchestrator.NetworkRequest{
		CIDR:                cidr,
		Name:                projectName,
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  publicSubnets,
		PrivateSubnetsCIDRs: privateSubnets}
	stack := infraUtil.Stack{
		StackName:   orchestrator.StackCreateNetworkName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault,
		DeployFunc:  request.CreateNetwork,
	}
	// Exec stack
	stackResult, err := infraUtil.ExecStack(stack)
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
	stack := infraUtil.Stack{
		StackName:   orchestrator.StackCreateNetworkName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	_, err := infraUtil.DestroyStack(stack)
	if err != nil {
		return err
	}
	logging.Debugf("VPC has been destroyed")
	return nil
}
