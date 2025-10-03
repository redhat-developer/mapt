package kind

import (
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
)

// TODO do some code to get this info from kind source code
type KindK8SImages struct {
	KindVersion string
	KindImage   string
}

var KindK8sVersions map[string]KindK8SImages = map[string]KindK8SImages{
	"v1.34": {"v0.30.0", "kindest/node:v1.34.0@sha256:7416a61b42b1662ca6ca89f02028ac133a309a2a30ba309614e8ec94d976dc5a"},
	"v1.33": {"v0.30.0", "kindest/node:v1.33.4@sha256:25a6018e48dfcaee478f4a59af81157a437f15e6e140bf103f85a2e7cd0cbbf2"},
	"v1.32": {"v0.30.0", "kindest/node:v1.32.8@sha256:abd489f042d2b644e2d033f5c2d900bc707798d075e8186cb65e3f1367a9d5a1"},
	"v1.31": {"v0.30.0", "kindest/node:v1.31.12@sha256:0f5cc49c5e73c0c2bb6e2df56e7df189240d83cf94edfa30946482eb08ec57d2"},
	"v1.30": {"v0.29.0", "kindest/node:v1.30.13@sha256:397209b3d947d154f6641f2d0ce8d473732bd91c87d9575ade99049aa33cd648"},
}

const (
	StackName = "stackKind"
	KindID    = "knd"
)

// TODO check if allow customize this, specially ingress related ports
var (
	PortHTTP  = 8888
	PortHTTPS = 9443
	PortAPI   = 6443
)

type KindArgs struct {
	Prefix            string
	ComputeRequest    *cr.ComputeRequestArgs
	Version           string
	Arch              string
	HostingPlace      string
	Spot              *spotTypes.SpotArgs
	Timeout           string
	ExtraPortMappings []PortMapping
}

type KindResults struct {
	Username   *string  `json:"username"`
	PrivateKey *string  `json:"private_key"`
	Host       *string  `json:"host"`
	Kubeconfig *string  `json:"kubeconfig"`
	SpotPrice  *float64 `json:"spot_price,omitempty"`
}
