package spot

import (
	"fmt"

	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/spot/aws"
	"github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util"
)

var amiRegexFedora = map[string]string{
	"amd64": "Fedora-Cloud-Base-%s*x86_64*",
	"arm64": "Fedora-Cloud-Base-%s*aarch64*",
}

var amiRegexRhel = map[string]string{
	"amd64": "RHEL-%s*-x86_64-*",
	"arm64": "RHEL-%s*-aarch64-*",
}

type SpotRequest struct {
	CPUs                  int32
	MemoryGib             int32
	Os                    string
	OSVersion             string
	Arch                  string
	NestedVirt            bool
	EvictionRateTolerance azure.EvictionRate
	MaxResults            int
}

type SpotPrice struct {
	InstanceType     string
	Price            float32
	Region           string
	AvailabilityZone string
}

func (sr *SpotRequest) getAwsProductDesc() string {
	switch sr.Os {
	case "windows":
		return "Windows"
	case "RHEL", "rhel":
		return "Red Hat Enterprise Linux"
	case "fedora":
		return "Linux/UNIX"
	default:
		return ""
	}
}

func (sr *SpotRequest) getAwsAMIName() string {
	switch sr.Os {
	case "fedora":
		return fmt.Sprintf(amiRegexFedora[sr.Arch], sr.OSVersion)
	case "RHEL", "rhel":
		return fmt.Sprintf(amiRegexRhel[sr.Arch], sr.OSVersion)
	case "windows":
		return "Windows_Server-2022-English-Full-HyperV-RHQE"
	default:
		return ""
	}
}

func (sr *SpotRequest) getAwsInstanceTypes() ([]string, error) {
	req := instancetypes.AwsInstanceRequest{
		CPUs:      sr.CPUs,
		MemoryGib: sr.MemoryGib,
		Arch: util.If(sr.Arch == "amd64",
			instancetypes.Amd64, instancetypes.Arm64),
		NestedVirt: sr.NestedVirt,
	}
	return req.GetMachineTypes()
}

func (sr *SpotRequest) getAzureInstanceTypes() ([]string, error) {
	req := instancetypes.AzureInstanceRequest{
		CPUs:      sr.CPUs,
		MemoryGib: sr.MemoryGib,
		Arch: util.If(sr.Arch == "amd64",
			instancetypes.Amd64, instancetypes.Arm64),
		NestedVirt: sr.NestedVirt,
	}
	return req.GetMachineTypes()
}

func (sr *SpotRequest) GetAwsLowestPrice() (SpotPrice, error) {
	vms, err := sr.getAwsInstanceTypes()
	if err != nil {
		return SpotPrice{}, err
	}

	arch := "x86_64"
	if sr.Arch == "arm64" {
		arch = "aarch64"
	}

	info, err := aws.BestSpotOptionInfo(sr.getAwsProductDesc(), vms, sr.getAwsAMIName(), arch)
	if err != nil {
		return SpotPrice{}, err
	}

	return SpotPrice{
		InstanceType:     info.InstanceType,
		Price:            float32(info.AVGPrice),
		Region:           info.Region,
		AvailabilityZone: info.AvailabilityZone,
	}, nil
}

func (sr *SpotRequest) getAzureOsType() string {
	switch sr.Os {
	case "fedora", "RHEL", "rhel", "ubuntu":
		return "linux"
	case "windows", "Windows":
		return "windows"
	default:
		return ""
	}
}

func (sr *SpotRequest) GetAzureLowestPrice() (SpotPrice, error) {
	vms, err := sr.getAzureInstanceTypes()
	if err != nil {
		return SpotPrice{}, nil
	}
	spr := azure.BestSpotChoiceRequest{
		VMTypes:               vms,
		OSType:                sr.getAzureOsType(),
		EvictionRateTolerance: azure.EvictionRate(sr.EvictionRateTolerance),
	}

	prices, err := azure.GetBestSpotChoice(spr)
	if err != nil {
		return SpotPrice{}, err
	}
	return SpotPrice{
		InstanceType: prices.VMType,
		Region:       prices.Location,
		Price:        float32(prices.Price),
	}, nil
}

// GetLowestPrice fetches prices of spot instances for all the supported
// providers and returns the results as a:  map[string]SpotPrice
// where map index key is the cloud provider name
func (sr *SpotRequest) GetLowestPrice() (map[string]SpotPrice, error) {
	var result = make(map[string]SpotPrice)
	azPrice, err := sr.GetAzureLowestPrice()
	if err != nil {
		return nil, err
	}
	result["azure"] = azPrice

	awsPrice, err := sr.GetAwsLowestPrice()
	if err != nil {
		return nil, err
	}
	result["aws"] = awsPrice
	return result, nil
}
