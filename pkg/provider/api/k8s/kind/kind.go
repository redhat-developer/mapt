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
	"v1.33": {"v0.29.0", "kindest/node:v1.33.1@sha256:050072256b9a903bd914c0b2866828150cb229cea0efe5892e2b644d5dd3b34f"},
	"v1.32": {"v0.29.0", "kindest/node:v1.32.5@sha256:e3b2327e3a5ab8c76f5ece68936e4cafaa82edf58486b769727ab0b3b97a5b0d"},
	"v1.31": {"v0.29.0", "kindest/node:v1.31.9@sha256:b94a3a6c06198d17f59cca8c6f486236fa05e2fb359cbd75dabbfc348a10b211"},
	"v1.30": {"v0.29.0", "kindest/node:v1.30.13@sha256:397209b3d947d154f6641f2d0ce8d473732bd91c87d9575ade99049aa33cd648"},
}

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
	Spot              *spotTypes.SpotArgs
	Timeout           string
	ExtraPortMappings []PortMapping
}

type KindResults struct {
	Username   string   `json:"username"`
	PrivateKey string   `json:"private_key"`
	Host       string   `json:"host"`
	Kubeconfig string   `json:"kubeconfig"`
	SpotPrice  *float64 `json:"spot_price,omitempty"`
}
