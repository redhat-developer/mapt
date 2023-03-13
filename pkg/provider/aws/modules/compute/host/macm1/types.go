package macm1

import (
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
)

type Request struct {
	compute.Request
	// Set the new password for the user, will be used to connect with vncviewer
	Password string
}

type Compute struct {
	compute.Compute
}
