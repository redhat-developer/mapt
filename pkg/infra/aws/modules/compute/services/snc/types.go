package snc

import "github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/host/rhel"

type SNCRequest struct {
	rhel.RHELRequest
}

type SNCCompute struct {
	rhel.RHELCompute
}
