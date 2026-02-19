package kind

var (
	diskSize int = 200

	// Official AMIs from Fedora use aarch64 format for arm64
	amiRegex = map[string]string{
		"x86_64": "Fedora-Cloud-Base-AmazonEC2.x86_64-4*",
		"arm64":  "Fedora-Cloud-Base-AmazonEC2.aarch64-4*",
	}
	// This is the ID for AMIS from https://fedoraproject.org/cloud
	amiOwner       = "125523088429"
	amiUserDefault = "fedora"
	amiProduct     = "Linux/UNIX"
)

func amiName(arch *string) string { return amiRegex[*arch] }
