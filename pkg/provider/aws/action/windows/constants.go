package windows

var (
	stackName = "stackWindowsBaremetal"

	awsWindowsDedicatedID     = "awd"
	diskSize              int = 200

	// This is based on a Custom AMI
	amiNameDefault  = "Windows_Server-2019-English-Full-HyperV-RHQE"
	amiOwnerDefault = "self"
	amiUserDefault  = "ec2-user"
	amiArchDefault  = "x86_64"

	// Custom non english ami
	amiLangNonEng        = "non-eng"
	amiNonEngNameDefault = "Windows_Server-2019-Spanish-Full-HyperV-RHQE"

	requiredInstanceTypes = []string{"c5.metal", "c5d.metal", "c5n.metal"}

	outputHost           = "awdHost"
	outputUsername       = "awdUsername"
	outputUserPassword   = "awdUserPassword"
	outputUserPrivateKey = "awdPrivatekey"
)
