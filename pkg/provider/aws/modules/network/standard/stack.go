package standard

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"

	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func Create(mCtxArgs *mc.ContextArgs, cidr string,
	azs, publicSubnets, privateSubnets, intraSubnets []string) error {
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	request := NetworkRequest{
		MCtx:                mCtx,
		CIDR:                cidr,
		Name:                mCtx.ProjectName(),
		AvailabilityZones:   azs,
		PublicSubnetsCIDRs:  publicSubnets,
		PrivateSubnetsCIDRs: privateSubnets,
		IntraSubnetsCIDRs:   intraSubnets,
		SingleNatGateway:    false}
	stack := manager.Stack{
		StackName:           StackCreateNetworkName,
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          request.Deployer,
	}
	// Exec stack
	stackResult, err := manager.UpStack(mCtx, stack)
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
	if err := r.validate(); err != nil {
		return err
	}
	_, err = r.CreateNetwork(ctx)
	return
}

func Destroy(mCtxArgs *mc.ContextArgs) (err error) {
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	stack := manager.Stack{
		StackName:           StackCreateNetworkName,
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials}
	err = manager.DestroyStack(mCtx, stack)
	if err == nil {
		logging.Debugf("VPC has been destroyed")
	}
	return
}
