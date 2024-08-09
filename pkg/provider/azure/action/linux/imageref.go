package linux

import (
	"fmt"
	"strings"
)

type OSType int

const (
	Ubuntu OSType = iota + 1
	RHEL
	Fedora
)

type imageReference struct {
	publisher string
	offer     string
	sku       string
}

var (
	defaultImageRefs = map[OSType]map[string]imageReference{
		RHEL: {
			"x86_64": {
				publisher: "RedHat",
				offer:     "RHEL",
				sku:       "%s_%s",
			},
			"arm64": {
				publisher: "RedHat",
				offer:     "rhel-arm64",
				sku:       "%s_%s-arm64",
			},
		},
		Ubuntu: {
			"x86_64": {
				publisher: "Canonical",
				offer:     "ubuntu-%s_%s-lts-daily",
				sku:       "server",
			},
		},
	}
)

// version should came in format X.Y (major.minor)
func getImageRef(osTarget OSType, arch string, version string) (*imageReference, error) {
	ir := defaultImageRefs[osTarget][arch]
	versions := strings.Split(version, ".")
	switch osTarget {
	case Ubuntu:
		return &imageReference{
			publisher: ir.publisher,
			offer:     fmt.Sprintf(ir.offer, versions[0], versions[1]),
			sku:       ir.sku,
		}, nil
	case RHEL:
		return &imageReference{
			publisher: ir.publisher,
			offer:     ir.offer,
			sku:       fmt.Sprintf(ir.sku, versions[0], versions[1]),
		}, nil
	}
	return nil, fmt.Errorf("os type not supported")

}
