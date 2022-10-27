package macm1

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
)

type MacM1Request struct {
	compute.Request
	// Set the new password for the user, will be used to connect with vncviewer
	Password string
}

type MacM1Compute struct {
	compute.Compute
}
