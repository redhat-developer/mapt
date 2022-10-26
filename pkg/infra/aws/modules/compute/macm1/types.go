package macm1

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type MacM1Request struct {
	compute.Request
}

type MacM1Resources struct {
	compute.Resources
}
