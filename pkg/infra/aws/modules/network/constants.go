package network

const (
	StackCreateNetworkName        string = "Manage-Network"
	StackCreateNetworkOutputVPCID string = "VPCID"
)

var (
	DefaultCIDRNetwork string = "10.0.0.0/16"

	DefaultCIDRPublicSubnets [3]string = [3]string{
		"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}
	DefaultLBIPs [3]string = [3]string{
		"10.0.1.15", "10.0.2.15", "10.0.3.15"}
	DefaultCIDRPrivateSubnets [3]string = [3]string{
		"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}
	DefaultCIDRIntraSubnets [3]string = [3]string{
		"10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24"}
	DefaultAvailabilityZones [3]string = [3]string{
		"us-east-1a", "us-east-1b", "us-east-1c"}
)
