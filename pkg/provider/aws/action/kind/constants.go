package kind

import "fmt"

var (
	diskSize int = 200

	// Official AMIs from Fedora use aarch64 format for arm64
	amiRegex = map[string]string{
		"x86_64": "Fedora-Cloud-Base-AmazonEC2.x86_64-%s-*",
		"arm64":  "Fedora-Cloud-Base-AmazonEC2.aarch64-%s-*",
	}
	// This is the ID for AMIS from https://fedoraproject.org/cloud
	amiOwner       = "125523088429"
	amiUserDefault = "fedora"
	amiProduct     = "Linux/UNIX"
)

// TODO check latest stable Fedora version
// for the time being we will use 41
func amiName(arch *string) string { return fmt.Sprintf(amiRegex[*arch], "41") }
