package allocation

import (
	"fmt"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/gcp/data"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type AllocationArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	Spot           *spotTypes.SpotArgs
}

type AllocationResult struct {
	Region       string
	Zone         string
	MachineTypes []string
	SpotPrice    *float64
}

func Allocation(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error) {
	if args.Spot != nil && args.Spot.Spot {
		return nil, fmt.Errorf("GCP spot allocation not yet implemented (Phase 2)")
	}
	return allocationOnDemand(mCtx, args)
}

func allocationOnDemand(mCtx *mc.Context, args *AllocationArgs) (*AllocationResult, error) {
	region := mCtx.TargetHostingPlace()

	zones, err := data.GetAvailabilityZones(mCtx.Context(), region)
	if err != nil {
		return nil, err
	}

	// Try random zones until we find one with matching machine types
	tried := make(map[string]bool)
	for len(tried) < len(zones) {
		idx := util.Random(len(zones)-1, 0)
		zone := zones[idx]
		if tried[zone] {
			continue
		}
		tried[zone] = true

		machineTypes, err := data.NewComputeSelector().Select(mCtx.Context(), args.ComputeRequest, zone)
		if err != nil {
			logging.Debugf("no matching machine types in zone %s: %v", zone, err)
			continue
		}

		if len(machineTypes) > 0 {
			logging.Debugf("allocated zone %s with machine types: %v", zone, machineTypes)
			return &AllocationResult{
				Region:       region,
				Zone:         zone,
				MachineTypes: machineTypes,
			}, nil
		}
	}

	return nil, fmt.Errorf("no zones in region %s have machine types matching the requested specs", region)
}
