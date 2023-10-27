package standard

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/vpc/subnet"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/services/vpc/vpc"
)

type NetworkRequest struct {
	CIDR                string
	Name                string
	Region              string
	AvailabilityZones   []string
	PublicSubnetsCIDRs  []string
	PrivateSubnetsCIDRs []string
	IntraSubnetsCIDRs   []string
	SingleNatGateway    bool
	PublicToIntra       *bool
}

type NetworkResources struct {
	VPCResources       *vpc.VPCResources
	AvailabilityZones  []string
	Region             string
	PublicSNResources  []*subnet.PublicSubnetResources
	PrivateSNResources []*subnet.PrivateSubnetResources
	IntraSNResources   []*subnet.PrivateSubnetResources
}
