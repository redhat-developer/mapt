package standard

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	"github.com/redhat-developer/mapt/pkg/provider/aws"

	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func Create(projectName, backedURL, cidr string,
	azs, publicSubnets, privateSubnets, intraSubnets []string) error {

	request := NetworkRequest{
		CIDR:                cidr,
		Name:                projectName,
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  publicSubnets,
		PrivateSubnetsCIDRs: privateSubnets,
		IntraSubnetsCIDRs:   intraSubnets,
		NatGatewayType:      ALL}
	stack := manager.Stack{
		StackName:           StackCreateNetworkName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          request.Deployer,
	}
	// Exec stack
	stackResult, err := manager.UpStack(stack)
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

func (r NetworkRequest) Deployer(ctx *pulumi.Context) (err error) {
	_, err = r.CreateNetwork(ctx)
	return
}

func Destroy(projectName, backedURL string) (err error) {
	stack := manager.Stack{
		StackName:           StackCreateNetworkName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		ProviderCredentials: aws.DefaultCredentials}
	err = manager.DestroyStack(stack)
	if err == nil {
		logging.Debugf("VPC has been destroyed")
	}
	return
}
