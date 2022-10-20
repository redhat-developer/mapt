package supportmatrix

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	OL_RHEL = SupportedHost{
		ID:                 olRHELID,
		Description:        "rhel machine supporting initialize / build openshift local",
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Red Hat Enterprise Linux",
		Requirements: &ec2.InstanceRequirementsWithMetadataRequest{
			ArchitectureTypes: aws.StringSlice([]string{"x86_64"}),
			InstanceRequirements: &ec2.InstanceRequirementsRequest{
				BareMetal: aws.String("required"),
				MemoryMiB: &ec2.MemoryMiBRequest{
					Max: aws.Int64(192000),
					Min: aws.Int64(192000),
				},
				VCpuCount: &ec2.VCpuCountRangeRequest{
					// Max: aws.Int64(192000),
					Min: aws.Int64(0),
				},
			},
		},
		Spot: true,
	}

	OL_Windows = SupportedHost{
		ID:                 olWindowsID,
		Description:        "windows machine supporting nested virtualization (start openshift local)",
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Windows",
		Requirements: &ec2.InstanceRequirementsWithMetadataRequest{
			ArchitectureTypes: aws.StringSlice([]string{"x86_64"}),
			InstanceRequirements: &ec2.InstanceRequirementsRequest{
				BareMetal: aws.String("required"),
				MemoryMiB: &ec2.MemoryMiBRequest{
					Max: aws.Int64(192000),
					Min: aws.Int64(192000),
				},
				VCpuCount: &ec2.VCpuCountRangeRequest{
					// Max: aws.Int64(192000),
					Min: aws.Int64(0),
				},
			},
		},
		Spot: true,
	}
)

func GetHost(id string) (*SupportedHost, error) {
	switch id {
	case olRHELID:
		return &OL_RHEL, nil
	case olWindowsID:
		return &OL_Windows, nil
	}
	return nil, fmt.Errorf("supported host id is not valid")
}
