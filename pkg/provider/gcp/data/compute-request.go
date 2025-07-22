package data

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/provider/gcp"
	hostingPlaces "github.com/redhat-developer/mapt/pkg/provider/util/hosting-places"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type ComputeSelector struct{}

func NewComputeSelector() *ComputeSelector { return &ComputeSelector{} }

func (c *ComputeSelector) Select(
	args *cr.ComputeRequestArgs) ([]string, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (c *ComputeSelector) SelectByHostingZone(
	args *cr.ComputeRequestArgs) (map[string][]string, error) {
	return machinesTypes(args)
}

func machinesTypes(args *cr.ComputeRequestArgs) (map[string][]string, error) {
	ctx := context.Background()
	machineTypesClient, err := compute.NewMachineTypesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create machineTypes client: %w", err)
	}
	defer func() {
		if err := machineTypesClient.Close(); err != nil {
			logging.Errorf("error closing gcp rest client")
		}
	}()
	zones, err := GetZones()
	if err != nil {
		return nil, err
	}
	results := hostingPlaces.RunOnHostingPlaces(
		zones,
		&machineTypesArgs{
			client:    machineTypesClient,
			projectID: gcp.GetProjectID(),
			ctx:       ctx,
		},
		machineTypesAsync)
	return results, fmt.Errorf("not implementedyet")
}

type machineTypesArgs struct {
	client    *compute.MachineTypesClient
	projectID string
	ctx       context.Context
}

func machineTypesAsync(z string, args *machineTypesArgs, c chan hostingPlaces.HostingPlaceData[[]string]) {
	req := computepb.ListMachineTypesRequest{
		Zone:    z,
		Project: args.projectID,
	}
	it := args.client.List(args.ctx, &req)
	var results []string
	var mt *computepb.MachineType
	var err error
	for {
		mt, err = it.Next()
		if err != nil {
			break
		}
		results = append(results, *mt.Name)
	}
	c <- hostingPlaces.HostingPlaceData[[]string]{
		Err:    err,
		Region: z,
		Value:  results}
}
