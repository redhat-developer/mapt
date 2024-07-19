package fedora

var (
	stackName            = "stackFedoraBaremetal"
	awsFedoraDedicatedID = "afd"

	diskSize int = 200

	// Official AMIs from Fedora use aarch64 format for arm64
	amiRegex = map[string]string{
		"x86_64": "Fedora-Cloud-Base-%s*x86_64*",
		"arm64":  "Fedora-Cloud-Base-%s*aarch64*",
	}
	// This is the ID for AMIS from https://fedoraproject.org/cloud
	amiOwner       = "125523088429"
	amiUserDefault = "fedora"

	supportedInstanceTypes = map[string][]string{
		"x86_64": {"c5.metal", "c5d.metal", "c5n.metal"},
		"arm64":  {"c7gd.metal", "c7gn.metal", "m6gd.metal"}}

	outputHost           = "afdHost"
	outputUsername       = "afdUsername"
	outputUserPrivateKey = "afdPrivatekey"
)
