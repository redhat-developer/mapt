package rhel

var (
	stackName          = "stackRHELBaremetal"
	awsRHELDedicatedID = "ard"

	diskSize int = 200

	amiProduct     = "Red Hat Enterprise Linux"
	amiRegex       = "RHEL-%s*-%s-*"
	amiUserDefault = "ec2-user"

	outputHost           = "ardHost"
	outputUsername       = "ardUsername"
	outputUserPrivateKey = "ardPrivatekey"
)
