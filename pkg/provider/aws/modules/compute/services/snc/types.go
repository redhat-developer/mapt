package snc

import "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute/host/rhel"

type SNCRequest struct {
	rhel.RHELRequest
}

type SNCCompute struct {
	rhel.RHELCompute
}
