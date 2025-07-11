package computerequest

import (
	"fmt"
)

const (
	Amd64 Arch = iota + 1
	Arm64

	MaxResults = 15 // maximum number of VM types to fetch
)

type Arch int

type Int32RangeFilter struct {
	UpperBound int32
	LowerBound int32
}

type ComputeRequestArgs struct {
	CPUs int32
	// CPUsRange       *Int32RangeFilter
	GPUs            int32
	GPUManufacturer string
	GPUModel        string
	MemoryGib       int32
	// MemoryRange     *Int32RangeFilter
	Arch       Arch
	NestedVirt bool
	// In case we want an specific type / size
	// we can set them directly
	ComputeSizes []string
}

type ComputeSelector interface {
	Select(args *ComputeRequestArgs) ([]string, error)
}

func (a Arch) String() string {
	switch a {
	case Amd64:
		return "x64"
	case Arm64:
		return "Arm64"
	}
	return ""
}

func Validate(cpus, memory int32, arch Arch) error {
	if cpus > 0 && memory > 0 && arch.String() != "" {
		return nil
	}
	return fmt.Errorf("invalid values for CPUs: %d, Memory: %d and Arch: %s", cpus, memory, arch)
}
