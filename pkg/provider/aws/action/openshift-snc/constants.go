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

	consoleURLRegex = "https://console-openshift-console.apps.%s.nip.io"

	outputHost           = "aosHost"
	outputUsername       = "aosUsername"
	outputUserPrivateKey = "aosPrivatekey"
	outputKubeconfig     = "aosKubeconfig"
	outputKubeAdminPass  = "aosKubeAdminPasss"
	outputDeveloperPass  = "aosDeveloperPass"

	commandCrcReadiness = "while [ ! -f /tmp/.crc-cluster-ready ]; do sleep 5; done"
	commandCaServiceRan = "sudo bash -c 'until oc get node --kubeconfig /opt/kubeconfig --context system:admin || oc get node --kubeconfig /opt/crc/kubeconfig --context system:admin; do sleep 5; done'"

	// portHTTP  = 80
	portHTTPS = 443
	portAPI   = 6443

	// SSM
	ocpPullSecretID = "ocppullsecretid"
	kapass          = "kapass"
	devpass         = "devpass"
)

func amiName(version, arch *string) string { return fmt.Sprintf(amiRegex, *version, *arch) }
