package kind

import "fmt"

var (
	stackName = "stackKind"
	awsKindID = "akd"

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

	outputHost           = "akdHost"
	outputUsername       = "akdUsername"
	outputUserPrivateKey = "akdPrivatekey"
	outputKubeconfig     = "akdKubeconfig"
)

// TODO do some code to get this info from kind source code
type kindK8SImages struct {
	kindVersion string
	KindImage   string
}

var KindK8sVersions map[string]kindK8SImages = map[string]kindK8SImages{
	"v1.32": {"v0.27.0", "kindest/node:v1.32.2@sha256:f226345927d7e348497136874b6d207e0b32cc52154ad8323129352923a3142f"},
	"v1.31": {"v0.27.0", "kindest/node:v1.31.6@sha256:28b7cbb993dfe093c76641a0c95807637213c9109b761f1d422c2400e22b8e87"},
	"v1.30": {"v0.27.0", "kindest/node:v1.30.10@sha256:4de75d0e82481ea846c0ed1de86328d821c1e6a6a91ac37bf804e5313670e507"},
	"v1.29": {"v0.27.0", "kindest/node:v1.29.14@sha256:8703bd94ee24e51b778d5556ae310c6c0fa67d761fae6379c8e0bb480e6fea29"},
}

// TODO check if allow customize this, specially ingress related ports
var (
	portHTTP  = 8888
	portHTTPS = 9443
	portAPI   = 6443
)

// TODO check latest stable Fedora version
// for the time being we will use 41
func amiName(arch *string) string { return fmt.Sprintf(amiRegex[*arch], "41") }
