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
			RegexName: "RHEL-8.7.0_HVM-*-x86_64*",
			// https://access.redhat.com/solutions/15356
			// Pattern with composition %s is major rhel version
			// RegexPattern: "RHEL-%s*-x86_64-*",
			DefaultUser: "ec2-user",
		},
	}

	OL_Fedora = SupportedHost{
		ID:                 olFedoraID,
		Description:        "fedora machine supporting nested virtualization (requirement to run openshift local)",
		Type:               Fedora,
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Linux/UNIX",
		Spot:               true,
		AMI: AMI{
			// https://alt.fedoraproject.org/en/cloud/
			// Pattern with composition %s is major fedora version
			RegexPattern: "Fedora-Cloud-Base-%s*",
			DefaultUser:  "fedora",
			// https://wiki.centos.org/Cloud/AWS
			Owner: "125523088429",
			Filters: map[string]string{
				"architecture": "x86_64"},
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

	S_SNC = SupportedHost{
		ID:                 sSNCID,
		Description:        "rhel machine with setup for build SNC",
		Type:               RHEL,
		InstaceTypes:       []string{"c5.metal", "c5d.metal", "c5n.metal"},
		ProductDescription: "Red Hat Enterprise Linux",
		Spot:               true,
		AMI: AMI{
			RegexName: "RHEL-9.2.0_HVM-*-x86_64*",
			// https://access.redhat.com/solutions/15356
			// Pattern with composition %s is major rhel version
			// RegexPattern: "RHEL-%s*-x86_64-*",
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
)

func GetHost(id string) (*SupportedHost, error) {
	switch id {
	case olRHELID:
		return &OL_RHEL, nil
	case olFedoraID:
		return &OL_Fedora, nil
	case olWindowsID:
		return &OL_Windows, nil
	case olWindowsNonEngID:
		return &OL_Windows_NonEng, nil
	case sSNCID:
		return &S_SNC, nil
	}
	return nil, fmt.Errorf("supported host id is not valid")
}
