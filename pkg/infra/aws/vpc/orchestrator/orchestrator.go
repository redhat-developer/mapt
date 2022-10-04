package orchestrator

import (
	vpc "github.com/adrianriobo/qenvs/pkg/infra/aws/vpc/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NetworkRequest struct {
	CIDR                string
	Name                string
	AvailabilityZones   []string
	PublicSubnetsCIDRs  []string
	PrivateSubnetsCIDRs []string
}

const (
	StackCreateNetworkName string = "Manage-Network"

	StackCreateNetworkOutputVPCID string = "VPCID"
)

func (r NetworkRequest) CreateNetwork(ctx *pulumi.Context) error {
	vpcRequest := vpc.VPCRequest{CIDR: r.CIDR, Name: r.Name}
	vpcResult, err := vpcRequest.CreateNetwork(ctx)
	if err != nil {
		return err
	}
	ctx.Export(StackCreateNetworkOutputVPCID, vpcResult.VPC.ID())
	return nil
}
