package data

import (
	"context"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"golang.org/x/exp/slices"
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
	f := filters(args)
	//nolint:staticcheck // following method is deprecated but no replacement yet
	details, err := instanceSelector.FilterVerbose(ctx, f)
	if err != nil {
		return nil, err
	}
	// The ec2-instance-selector library does not always honor CPUArchitecture
	// when GPUManufacturer is set, so we post-filter by arch.
	wantArch := *arch(args.Arch)
	var result []string
	for _, d := range details {
		if slices.Contains(d.ProcessorInfo.SupportedArchitectures, wantArch) {
			result = append(result, string(d.InstanceType))
		}
	}
	return result, nil
}

func filters(args *computerequest.ComputeRequestArgs) (f selector.Filters) {
	if args.NestedVirt {
		f.CPUArchitecture = arch(args.Arch)
		f.BareMetal = &args.NestedVirt
		return
	}
	if len(args.GPUManufacturer) > 0 {
		f.GPUManufacturer = &args.GPUManufacturer
		f.CPUArchitecture = arch(args.Arch)
		if args.CPUs > 0 {
			upperBound := args.MaxCPUs
			if upperBound == 0 {
				upperBound = args.CPUs
			}
			f.VCpusRange = &selector.Int32RangeFilter{
				LowerBound: args.CPUs,
				UpperBound: upperBound,
			}
		}
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

