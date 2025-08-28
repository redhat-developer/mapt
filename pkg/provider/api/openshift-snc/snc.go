package openshiftsnc

import (
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/api/spot"
)

var (
	ConsoleURLRegex = "https://console-openshift-console.apps.%s.nip.io"

	OutputHost           = "aosHost"
	OutputUsername       = "aosUsername"
	OutputUserPrivateKey = "aosPrivatekey"
	OutputKubeconfig     = "aosKubeconfig"
	OutputKubeAdminPass  = "aosKubeAdminPasss"
	OutputDeveloperPass  = "aosDeveloperPass"

	CommandReadiness    = "while [ ! -f /tmp/.crc-cluster-ready ]; do sleep 5; done"
	CommandCaServiceRan = "sudo bash -c 'until oc get node --kubeconfig /opt/kubeconfig --context system:admin || oc get node --kubeconfig /opt/crc/kubeconfig --context system:admin; do sleep 5; done'"

	// portHTTP  = 80
	PortHTTPS = 443
	PortAPI   = 6443
)

type OpenshiftSNCArgs struct {
	Location       string
	Prefix         string
	ComputeRequest *cr.ComputeRequestArgs
	Version        string
	Arch           string
	PullSecretFile string
	Spot           *spot.SpotArgs
	Timeout        string
}

type OpenshiftSncResultsMetadata struct {
	Username      string   `json:"username"`
	PrivateKey    string   `json:"private_key"`
	Host          string   `json:"host"`
	Kubeconfig    string   `json:"kubeconfig"`
	KubeadminPass string   `json:"kubeadmin_pass"`
	SpotPrice     *float64 `json:"spot_price,omitempty"`
	ConsoleUrl    string   `json:"console_url,omitempty"`
}
