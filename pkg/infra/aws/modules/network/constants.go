package network

const (
	StackCreateNetworkName        string = "Manage-Network"
	StackCreateNetworkOutputVPCID string = "VPCID"
)

var (
	cidrNetowrk        string    = "10.0.0.0/16"
	cidrPublicSubnets  [3]string = [3]string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}
	cidrPrivateSubnets [3]string = [3]string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}
	cidrIntraSubnets   [3]string = [3]string{"10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24"}
)
