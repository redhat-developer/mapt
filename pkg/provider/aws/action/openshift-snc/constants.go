package openshiftsnc

import "fmt"

var (
	stackName   = "stackOpenshiftSNC"
	awsOCPSNCID = "aos"

	diskSize int = 200

	// This is managed by https://github.com/devtools-qe-incubator/cloud-importer
	amiProductDescription = "Linux/UNIX"
	// amiProductDescription = "Red Hat Enterprise Linux"
	amiRegex       = "openshift-local-%s-%s"
	amiUserDefault = "core"
	amiOwner       = "391597328979"
	// amiOriginRegion       = "us-east-1"

	consoleURLRegex = "https://console-openshift-console.apps.%s.nip.io"

	outputHost           = "aosHost"
	outputUsername       = "aosUsername"
	outputUserPrivateKey = "aosPrivatekey"
	outputKubeconfig     = "aosKubeconfig"
	outputKubeAdminPass  = "aosKubeAdminPasss"
	outputDeveloperPass  = "aosDeveloperPass"

	commandReadiness = "while [ ! -f /tmp/.crc-cluster-ready ]; do sleep 5; done"

	// portHTTP  = 80
	portHTTPS = 443
	portAPI   = 6443

	// SSM
	ocpPullSecretID = "ocppullsecretid"
	cacertID        = "cacertid"
	kapass          = "kapass"
	devpass         = "devpass"
)

func amiName(version, arch *string) string { return fmt.Sprintf(amiRegex, *version, *arch) }
