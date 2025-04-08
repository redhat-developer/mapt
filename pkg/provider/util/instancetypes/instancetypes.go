package instancetypes

import "fmt"

type InstanceRequest interface {
	GetMachineTypes() ([]string, error)
}

type arch int

func (a arch) String() string {
	switch a {
	case Amd64:
		return "x64"
	case Arm64:
		return "Arm64"
	}
	return ""
}

const (
	Amd64 arch = iota + 1
	Arm64
)

const maxResults = 15 // maximum number of VM types to fetch

func validate(cpus, memory int32, arch arch) error {
	if cpus > 0 && memory > 0 && arch.String() != "" {
		return nil
	}
	return fmt.Errorf("invalid values for CPUs: %d, Memory: %d and Arch: %s", cpus, memory, arch)
}
