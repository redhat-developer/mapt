package kind

var (
	diskSize int = 200

	// Official AMIs from Fedora use aarch64 format for arm64.
	// Pinned to Fedora 41 to ensure consistent cloud-init behavior
	// and avoid regressions from new Fedora releases.
	// See https://github.com/redhat-developer/mapt/issues/919
	amiRegex = map[string]string{
		"x86_64": "Fedora-Cloud-Base-AmazonEC2.x86_64-41.*",
		"arm64":  "Fedora-Cloud-Base-AmazonEC2.aarch64-41.*",
	}
	// This is the ID for AMIS from https://fedoraproject.org/cloud
	amiOwner       = "125523088429"
	amiUserDefault = "fedora"
	amiProduct     = "Linux/UNIX"
)

func amiName(arch *string) string { return amiRegex[*arch] }
