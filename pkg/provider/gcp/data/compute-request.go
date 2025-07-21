package data

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type ComputeSelector struct{}

func NewComputeSelector() *ComputeSelector { return &ComputeSelector{} }

func (c *ComputeSelector) Select(
	args *cr.ComputeRequestArgs) ([]string, error) {
	return machinesTypes(args)
}

func (c *ComputeSelector) SelectByHostingZone(
	args *cr.ComputeRequestArgs) (map[string][]string, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func machinesTypes(args *cr.ComputeRequestArgs) ([]string, error) {
	machineTypesClient, err := compute.NewMachineTypesRESTClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create machineTypes client: %w", err)
	}
	defer func() {
		if err := machineTypesClient.Close(); err != nil {
			logging.Errorf("error closing gcp rest client")
		}
	}()

	// reqList := &computepb.ListMachineTypesRequest{
	// 	Project: projectID,
	// 	Zone:    zone,
	// }

	return nil, fmt.Errorf("not implementedyet")
}
