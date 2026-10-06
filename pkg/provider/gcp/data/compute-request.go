package data

import (
	"context"
	"fmt"
	"sort"
	"strings"

	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	gcpAPI "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type ComputeSelector struct{}

func NewComputeSelector() *ComputeSelector {
	return &ComputeSelector{}
}

func (s *ComputeSelector) Select(ctx context.Context, args *cr.ComputeRequestArgs, zone string) ([]string, error) {
	if len(args.ComputeSizes) > 0 {
		return args.ComputeSizes, nil
	}

	svc, err := gcpAPI.NewService(ctx, option.WithScopes(gcpAPI.ComputeReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("error creating GCP compute service: %v", err)
	}

	project, err := getProject(ctx)
	if err != nil {
		return nil, err
	}

	mtList, err := svc.MachineTypes.List(project, zone).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing machine types in zone %s: %v", zone, err)
	}

	var matches []*gcpAPI.MachineType
	for _, mt := range mtList.Items {
		if !matchesCPU(mt, args) {
			continue
		}
		if !matchesMemory(mt, args) {
			continue
		}
		if !matchesFamily(mt, args) {
			continue
		}
		matches = append(matches, mt)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no GCP machine types match the requested specs (cpus>=%d, memory>=%dGiB) in zone %s",
			args.CPUs, args.MemoryGib, zone)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].GuestCpus < matches[j].GuestCpus
	})

	limit := cr.MaxResults
	if len(matches) < limit {
		limit = len(matches)
	}

	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = matches[i].Name
	}

	logging.Debugf("GCP compute selector found %d matching machine types in %s, returning top %d", len(matches), zone, limit)
	return result, nil
}

func matchesCPU(mt *gcpAPI.MachineType, args *cr.ComputeRequestArgs) bool {
	if mt.GuestCpus < int64(args.CPUs) {
		return false
	}
	if args.MaxCPUs > 0 && mt.GuestCpus > int64(args.MaxCPUs) {
		return false
	}
	return true
}

func matchesMemory(mt *gcpAPI.MachineType, args *cr.ComputeRequestArgs) bool {
	requiredMB := int64(args.MemoryGib) * 1024
	return mt.MemoryMb >= requiredMB
}

func matchesFamily(mt *gcpAPI.MachineType, args *cr.ComputeRequestArgs) bool {
	if len(args.ComputeFamilies) == 0 {
		return true
	}
	for _, family := range args.ComputeFamilies {
		if strings.HasPrefix(mt.Name, family+"-") {
			return true
		}
	}
	return false
}
