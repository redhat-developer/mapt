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
	"v1.35": {"v0.31.0", "kindest/node:v1.35.0@sha256:452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f"},
	"v1.34": {"v0.31.0", "kindest/node:v1.34.3@sha256:08497ee19eace7b4b5348db5c6a1591d7752b164530a36f855cb0f2bdcbadd48"},
	"v1.33": {"v0.31.0", "kindest/node:v1.33.7@sha256:d26ef333bdb2cbe9862a0f7c3803ecc7b4303d8cea8e814b481b09949d353040"},
	"v1.32": {"v0.31.0", "kindest/node:v1.32.11@sha256:5fc52d52a7b9574015299724bd68f183702956aa4a2116ae75a63cb574b35af8"},
	"v1.31": {"v0.31.0", "kindest/node:v1.31.14@sha256:6f86cf509dbb42767b6e79debc3f2c32e4ee01386f0489b3b2be24b0a55aac2b"},
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
