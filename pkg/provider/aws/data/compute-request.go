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
	ctx context.Context, args *computerequest.ComputeRequestArgs) ([]string, error) {
	return getInstanceTypes(ctx, args)
}

func getInstanceTypes(ctx context.Context, args *computerequest.ComputeRequestArgs) ([]string, error) {
	// if err := validate(r.CPUs, r.MemoryGib, r.Arch); err != nil {
	// 	return nil, err
	// }
	cfg, err := getGlobalConfig(ctx)
	if err != nil {
		return nil, err
	}
	instanceSelector, err := selector.New(
		ctx,
		cfg)
	if err != nil {
		return nil, err
	}
	//nolint:staticcheck // following method is deprecated but no replacement yet
	instanceTypesSlice, err := instanceSelector.Filter(
		ctx,
		filters(args))
	if err != nil {
		return nil, err
	}
	return instanceTypesSlice, nil
}

func filters(args *computerequest.ComputeRequestArgs) (f selector.Filters) {
	if args.NestedVirt {
		f.CPUArchitecture = arch(args.Arch)
		f.BareMetal = &args.NestedVirt
		return
	}
	if len(args.GPUManufacturer) > 0 {
		f.GPUManufacturer = &args.GPUManufacturer
		// filters.GpuMemoryRange = &selector.ByteQuantityRangeFilter{
		// 	LowerBound: mbq,
		// 	UpperBound: mbq,
		// }
	} else {
		f.VCpusRange = &selector.Int32RangeFilter{
			LowerBound: args.CPUs,
			UpperBound: args.CPUs,
		}
		mbq := bytequantity.FromGiB(
			uint64(args.MemoryGib))
		f.MemoryRange = &selector.ByteQuantityRangeFilter{
			LowerBound: mbq,
			UpperBound: mbq,
		}
		f.CPUArchitecture = arch(args.Arch)
		maxResults := computerequest.MaxResults
		f.MaxResults = &maxResults
	}
	return
}

func arch(ca computerequest.Arch) *ec2types.ArchitectureType {
	arch := ec2types.ArchitectureTypeX8664
	if ca == computerequest.Arm64 {
		arch = ec2types.ArchitectureTypeArm64
	}
	return &arch
}
