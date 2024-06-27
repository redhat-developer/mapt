package rhel

var (
	stackName          = "stackRHELBaremetal"
	awsRHELDedicatedID = "ard"

	diskSize int = 200

	amiProductDescription = "Red Hat Enterprise Linux"
	amiRegex              = "RHEL-%s*-%s-*"
	amiUserDefault        = "ec2-user"

	supportedInstanceTypes = map[string][]string{
		"x86_64": {"c5.metal", "c5d.metal", "c5n.metal"},
		"arm64":  {"c7gd.metal", "c7gn.metal", "m6gd.metal"}}

	outputHost           = "ardHost"
	outputUsername       = "ardUsername"
	outputUserPrivateKey = "ardPrivatekey"
)
