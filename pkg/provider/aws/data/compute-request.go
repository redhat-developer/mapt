package data

import (
	"context"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
)

type ComputeSelector struct{}

func NewComputeSelector() *ComputeSelector { return &ComputeSelector{} }

func (c *ComputeSelector) Select(
	args *computerequest.ComputeRequestArgs) ([]string, error) {
	return getInstanceTypes(args)
}

func getInstanceTypes(args *computerequest.ComputeRequestArgs) ([]string, error) {
	// if err := validate(r.CPUs, r.MemoryGib, r.Arch); err != nil {
	// 	return nil, err
	// }
	cfg, err := getGlobalConfig()
	if err != nil {
		return nil, err
	}
	instanceSelector, err := selector.New(
		context.Background(),
		cfg)
	if err != nil {
		return nil, err
	}
	var filters selector.Filters
	if len(args.GPUManufacturer) > 0 {
		filters.GPUManufacturer = &args.GPUManufacturer
		// filters.GpuMemoryRange = &selector.ByteQuantityRangeFilter{
		// 	LowerBound: mbq,
		// 	UpperBound: mbq,
		// }
	} else {
		filters.VCpusRange = &selector.Int32RangeFilter{
			LowerBound: args.CPUs,
			UpperBound: args.CPUs,
		}
		mbq := bytequantity.FromGiB(
			uint64(args.MemoryGib))
		filters.MemoryRange = &selector.ByteQuantityRangeFilter{
			LowerBound: mbq,
			UpperBound: mbq,
		}
		arch := ec2types.ArchitectureTypeX8664
		if args.Arch == computerequest.Arm64 {
			arch = ec2types.ArchitectureTypeArm64
		}
		filters.CPUArchitecture = &arch
		maxResults := computerequest.MaxResults
		filters.MaxResults = &maxResults
	}
	//nolint:staticcheck // following method is deprecated but no replacement yet
	instanceTypesSlice, err := instanceSelector.Filter(
		context.Background(),
		filters)
	if err != nil {
		return nil, err
	}
	return instanceTypesSlice, nil
}
