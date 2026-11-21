package data

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetImageRefInvalidVersion(t *testing.T) {
	for _, osTarget := range []OSType{Ubuntu, RHEL} {
		for _, version := range []string{"", "24", ".04", "24.", "."} {
			t.Run(fmt.Sprintf("os-%d/%q", osTarget, version), func(t *testing.T) {
				image, err := GetImageRef(osTarget, "x86_64", version)
				if err == nil || !strings.Contains(err.Error(), "expected X.Y") {
					t.Fatalf("expected a version-format error, got image=%+v, err=%v", image, err)
				}
				if image != nil {
					t.Errorf("expected no image for invalid version, got %+v", image)
				}
			})
		}
	}
}

func TestGetImageRefValidVersion(t *testing.T) {
	for _, tt := range []struct {
		name          string
		os            OSType
		arch, version string
		want          ImageReference
	}{
		{"ubuntu", Ubuntu, "x86_64", "24.04", ImageReference{Publisher: "Canonical", Offer: "ubuntu-24_04-lts-daily", Sku: "server"}},
		{"rhel", RHEL, "x86_64", "9.7", ImageReference{Publisher: "RedHat", Offer: "RHEL", Sku: "9_7"}},
		{"rhel-arm64", RHEL, "arm64", "9.7", ImageReference{Publisher: "RedHat", Offer: "rhel-arm64", Sku: "9_7-arm64"}},
		{"fedora-major", Fedora, "x86_64", "43", ImageReference{CommunityImageID: fedoraImageGalleryBase + "Fedora-Cloud-43-x64/Versions/latest"}},
		{"fedora-arm64", Fedora, "arm64", "43", ImageReference{CommunityImageID: fedoraImageGalleryBase + "Fedora-Cloud-43-Arm64/Versions/latest"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			image, err := GetImageRef(tt.os, tt.arch, tt.version)
			if err != nil {
				t.Fatal(err)
			}
			if *image != tt.want {
				t.Errorf("got %+v, want %+v", *image, tt.want)
			}
		})
	}
}
