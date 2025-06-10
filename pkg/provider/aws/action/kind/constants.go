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
	"v1.33": {"v0.29.0", "kindest/node:v1.33.1@sha256:050072256b9a903bd914c0b2866828150cb229cea0efe5892e2b644d5dd3b34f"},
	"v1.32": {"v0.29.0", "kindest/node:v1.32.5@sha256:e3b2327e3a5ab8c76f5ece68936e4cafaa82edf58486b769727ab0b3b97a5b0d"},
	"v1.31": {"v0.29.0", "kindest/node:v1.31.9@sha256:b94a3a6c06198d17f59cca8c6f486236fa05e2fb359cbd75dabbfc348a10b211"},
	"v1.30": {"v0.29.0", "kindest/node:v1.30.13@sha256:397209b3d947d154f6641f2d0ce8d473732bd91c87d9575ade99049aa33cd648"},
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
