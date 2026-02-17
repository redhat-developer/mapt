package snc

import (
	"fmt"

	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/util"
	"golang.org/x/mod/semver"
)

var (
	StackName = "stackOpenshiftSNC"
	OCPSNCID  = "snc"

	consoleURLRegex = "https://console-openshift-console.apps.%s.nip.io"

	OutputHost           = "aosHost"
	OutputUsername       = "aosUsername"
	OutputUserPrivateKey = "aosPrivatekey"
	OutputKubeconfig     = "aosKubeconfig"
	OutputKubeAdminPass  = "aosKubeAdminPasss"
	OutputDeveloperPass  = "aosDeveloperPass"

	PortHTTPS = 443
	PortAPI   = 6443
)

var (
	ClientKubeconfigPath       = "/opt/crc/kubeconfig"
	ContextAdminStarterVersion = "4.20.8"
	contextAdmin               = "admin"
	contextSystemAdmin         = "system:admin"
	CommandKubeconfigExists    = fmt.Sprintf("while [ ! -f %s ]; do sleep 5; done", ClientKubeconfigPath)
	CommandCrcReadiness        = "while [ ! -f /tmp/.crc-cluster-ready ]; do sleep 5; done"
	commandCaServiceRan        = "sudo bash -c 'until oc get node --kubeconfig /opt/kubeconfig --context %s || oc get node --kubeconfig /opt/crc/kubeconfig --context system:admin; do sleep 5; done'"
)

func CommandCaServiceRan(version string) string {
	return fmt.Sprintf(commandCaServiceRan, util.If(semver.Compare(version, ContextAdminStarterVersion) < 0, contextSystemAdmin, contextAdmin))
}

type SNCArgs struct {
	Prefix                  string
	ComputeRequest          *cr.ComputeRequestArgs
	Version                 string
	DisableClusterReadiness bool
	Arch                    string
	PullSecretFile          string
	Spot                    *spotTypes.SpotArgs
	Timeout                 string
}

type SNCResults struct {
	Username      string   `json:"username"`
	PrivateKey    string   `json:"private_key"`
	Host          string   `json:"host"`
	Kubeconfig    string   `json:"kubeconfig"`
	KubeadminPass string   `json:"kubeadmin_pass"`
	SpotPrice     *float64 `json:"spot_price,omitempty"`
	ConsoleUrl    string   `json:"console_url,omitempty"`
}
