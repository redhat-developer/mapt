package rhel

var (
	stackName          = "stackRHELBaremetal"
	awsRHELDedicatedID = "ard"

	diskSize int = 200

	amiRegex       = "RHEL-%s*-x86_64-*"
	amiUserDefault = "ec2-user"

	requiredInstanceTypes = []string{"c5.metal", "c5d.metal", "c5n.metal"}

	outputHost           = "ardHost"
	outputUsername       = "ardUsername"
	outputUserPrivateKey = "ardPrivatekey"
)
