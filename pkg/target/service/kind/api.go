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
	"v1.37": {"v0.33.0", "kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5"},
	"v1.36": {"v0.33.0", "kindest/node:v1.36.4@sha256:099e049362a1526b2db71494e1947aae99bd16290d7c895f2b7ea312e3cbfaed"},
	"v1.35": {"v0.33.0", "kindest/node:v1.35.8@sha256:07b2536e30b803ed61d1677a79df6115f798ce64c80f9e22f6ed45afd09323c0"},
	"v1.34": {"v0.33.0", "kindest/node:v1.34.11@sha256:44e222ee2132dab25ff87301682f89eb82c7880ea3a1bf543bfe9708fd08d67d"},
	"v1.33": {"v0.32.0", "kindest/node:v1.33.12@sha256:3f5c8443c620245e4d355cfe09e96a91ead32ceaa569d3f1ca9edf0cb2fe2ff4"},
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
	ServiceEndpoints         []string
	ExtraPortMappings []PortMapping
}

type KindResults struct {
	Username   *string  `json:"username"`
	PrivateKey *string  `json:"private_key"`
	Host       *string  `json:"host"`
	Kubeconfig *string  `json:"kubeconfig"`
	SpotPrice  *float64 `json:"spot_price,omitempty"`
}
