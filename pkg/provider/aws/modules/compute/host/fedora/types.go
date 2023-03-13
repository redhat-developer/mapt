package fedora

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
)

type Request struct {
	VersionMajor string
	compute.Request
}

type Compute struct {
	compute.Compute
}
