package openspotng

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type OpenspotNGRequest struct {
	compute.Request
	OCPPullSecretFilePath string
	DeveloperPassword     string
	KubeadminPassword     string
	RedHatPassword        string
}

type OpenspotNGCompute struct {
	compute.Compute
}
