package openshiftsnc

import "fmt"

var (
	stackName   = "stackOpenshiftSNC"
	awsOCPSNCID = "aos"

	diskSize int = 200

	// This is managed by https://github.com/devtools-qe-incubator/cloud-importer
	amiProduct = "Linux/UNIX"
	// amiProductDescription = "Red Hat Enterprise Linux"
	amiRegex       = "openshift-local-%s-%s-*"
	amiUserDefault = "core"
	amiOwner       = "391597328979"
	// amiOriginRegion       = "us-east-1"

	// SSM
	ocpPullSecretID = "ocppullsecretid"
	kapass          = "kapass"
	devpass         = "devpass"
)

func amiName(version, arch *string) string { return fmt.Sprintf(amiRegex, *version, *arch) }
