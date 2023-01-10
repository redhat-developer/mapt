package supportmatrix

import (
	"fmt"
)

var (
	OL_RHEL = SupportedHost{
		ID:                 olRHELID,
		Description:        "rhel machine supporting nested virtualization (requirement to run openshift local)",
		Type:               RHEL,
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Red Hat Enterprise Linux",
		Spot:               true,
		AMI: AMI{
			// https://access.redhat.com/solutions/15356
			// Pattern with composition %s is major rhel version
			RegexPattern: "RHEL-%s*-x86_64-*",
			DefaultUser:  "ec2-user",
		},
	}

	OL_Windows = SupportedHost{
		ID:                 olWindowsID,
		Description:        "windows machine supporting nested virtualization (start openshift local)",
		Type:               Windows,
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Windows",
		Spot:               true,
		AMI: AMI{
			// RegexName:   "Windows_Server-2019-English-Full-HyperV*",
			RegexName:   "Windows_Server-2019-English-Full-OCPL-*-RHQE*",
			DefaultUser: "ec2-user",
			Owner:       OwnerSelf,
		},
	}

	OL_Windows_NonEng = SupportedHost{
		ID:                 olWindowsID,
		Description:        "non english windows machine supporting nested virtualization (start openshift local)",
		Type:               Windows,
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Windows",
		Spot:               true,
		AMI: AMI{
			// add param to set openshift local version
			// RegexPattern: "Windows_Server-2019-Spanish-Full-OCPL-%s-RHQE*",
			RegexName:   "Windows_Server-2019-Spanish-Full-OCPL-*-RHQE*",
			DefaultUser: "ec2-user",
			Owner:       OwnerSelf,
		},
	}

	G_MAC_M1 = SupportedHost{
		ID:           gMacOSM1ID,
		Description:  "mac machine with m1 chip arm64 arch",
		Type:         MacM1,
		InstaceTypes: []string{"mac2.metal"},
		Spot:         false,
		AMI: AMI{
			RegexName:   "amzn-ec2-macos-12*",
			DefaultUser: "ec2-user",
			Owner:       "628277914472",
			Filters: map[string]string{
				"architecture": "arm64_mac"},
		},
	}

	S_SNC = SupportedHost{
		ID:                 sSNCID,
		Description:        "rhel machine with setup for build SNC",
		Type:               RHEL,
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Red Hat Enterprise Linux",
		Spot:               true,
		AMI: AMI{
			// https://access.redhat.com/solutions/15356
			// Pattern with composition %s is major rhel version
			RegexPattern: "RHEL-%s*-x86_64-*",
			DefaultUser:  "ec2-user",
		},
	}

	S_BASTION = SupportedHost{
		ID:           sBastionID,
		Description:  "bastion host to access hosts on private subnets",
		InstaceTypes: []string{"t2.small"},
		Spot:         false,
		AMI: AMI{
			RegexName:   "amzn-ami-hvm-*-x86_64-ebs",
			DefaultUser: "ec2-user",
		},
	}

	// https://github.com/ptcodes/proxy-server-with-terraform/blob/master/main.tf
	S_PROXY = SupportedHost{
		ID:           sProxyID,
		Description:  "proxy host to control network http access from hosts",
		InstaceTypes: []string{"t2.small"},
		Spot:         false,
		AMI: AMI{
			RegexName:   "amzn-ami-hvm-*-x86_64-ebs",
			DefaultUser: "ec2-user",
		},
	}

	S_OPENSPOTNG = SupportedHost{
		ID:          sOpenspotNG,
		Description: "openspotng cluster offering",
		Type:        RHEL,
		// InstaceTypes: []string{"c6in.2xlarge", "c7g.2xlarge", "c6g.2xlarge"},
		InstaceTypes:       []string{"c6in.2xlarge"},
		ProductDescription: "Linux/UNIX",
		// Spot:               true,
		Spot: false,
		FixedAMI: &FixedAMI{
			AvailavilityZone: "us-west-2b",
			Region:           "us-west-2",
		},
		AMI: AMI{
			RegexName:   "CRC Openshift*",
			DefaultUser: "core",
			// Owner:           OwnerSelf,
			Owner: "452958939641",
			//
			AMITargetName:   "CRC Openshift",
			AMISourceID:     "ami-0569ce8a44f2351be",
			AMISourceRegion: "us-west-2",
		},
	}
)

func GetHost(id string) (*SupportedHost, error) {
	switch id {
	case olRHELID:
		return &OL_RHEL, nil
	case olWindowsID:
		return &OL_Windows, nil
	case olWindowsNonEngID:
		return &OL_Windows_NonEng, nil
	case gMacOSM1ID:
		return &G_MAC_M1, nil
	case sSNCID:
		return &S_SNC, nil
	case sOpenspotNG:
		return &S_OPENSPOTNG, nil
	}
	return nil, fmt.Errorf("supported host id is not valid")
}
