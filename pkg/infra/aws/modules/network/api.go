package network

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"

	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

func CreateNetwork(projectName, backedURL, cidr string,
	azs, publicSubnets, privateSubnets, intraSubnets []string) error {

	request := NetworkRequest{
		CIDR:                cidr,
		Name:                projectName,
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  publicSubnets,
		PrivateSubnetsCIDRs: privateSubnets,
		IntraSubnetsCIDRs:   intraSubnets,
		SingleNatGateway:    false}
	stack := utilInfra.Stack{
		StackName:   StackCreateNetworkName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault,
		DeployFunc:  request.Deployer,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return err
	}
	vpcID, ok := stackResult.Outputs[StackCreateNetworkOutputVPCID].Value.(string)
	if !ok {
		return fmt.Errorf("error getting vpc id")
	}
	logging.Debugf("VPC has been created with id: %s", vpcID)
	return nil
}

func DestroyNetwork(projectName, backedURL string) (err error) {
	stack := utilInfra.Stack{
		StackName:   StackCreateNetworkName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	err = utilInfra.DestroyStack(stack)
	if err == nil {
		logging.Debugf("VPC has been destroyed")
	}
	return
}
