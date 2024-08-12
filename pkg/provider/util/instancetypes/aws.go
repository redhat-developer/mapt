package instancetypes

import (
	"context"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AwsInstanceRequest struct {
	CPUs       int32
	MemoryGib  int32
	Arch       arch
	NestedVirt bool
}

func (r *AwsInstanceRequest) GetMachineTypes() ([]string, error) {
	if err := validate(r.CPUs, r.MemoryGib, r.Arch); err != nil {
		return nil, err
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	instanceSelector, err := selector.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	vcpusRange := selector.Int32RangeFilter{
		LowerBound: r.CPUs,
		UpperBound: r.CPUs,
	}
	memoryRange := selector.ByteQuantityRangeFilter{
		LowerBound: bytequantity.FromGiB(uint64(r.MemoryGib)),
		UpperBound: bytequantity.FromGiB(uint64(r.MemoryGib)),
	}

	arch := ec2types.ArchitectureTypeX8664
	if r.Arch == Arm64 {
		arch = ec2types.ArchitectureTypeArm64
	}

	maxResults := maxResults

	filters := selector.Filters{
		VCpusRange:      &vcpusRange,
		MemoryRange:     &memoryRange,
		CPUArchitecture: &arch,
		MaxResults:      &maxResults,
		BareMetal:       &r.NestedVirt,
	}

	instanceTypesSlice, err := instanceSelector.Filter(ctx, filters)
	if err != nil {
		return nil, err
	}
	return instanceTypesSlice, nil
}
