package constants

import (
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/mac"
)

const (
	MACRequestCmd      = "request"
	MACRRequestCmdDesc = "request mac machine"
	MACReleaseCmd      = "release"
	MACReleaseCmdDesc  = "release mac machine"

	MACArch              string = "arch"
	MACArchDesc          string = "MAC architecture allowed values x86, m1, m2"
	MACArchDefault       string = mac.DefaultArch
	MACOSVersion         string = "version"
	MACOSVersionDesc     string = "MACos operating system version 11, 12 on x86 and m1/m2; 13, 14, 15 on all archs"
	MACOSVersionDefault  string = mac.DefaultOSVersion
	MACFixedLocation     string = "fixed-location"
	MACFixedLocationDesc string = "if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)"
	MACDHID              string = "dedicated-host-id"
	MACDHIDDesc          string = "id for the dedicated host"

	Spot            string = "spot"
	SpotDesc        string = "if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)"
)

func MACArchAsCirrusArch(arch string) *cirrus.Arch {
	switch arch {
	case "x86":
		return &cirrus.Amd64
	}
	return &cirrus.Arm64
}

func MACArchAsGithubArch(arch string) *github.Arch {
	switch arch {
	case "x86_64":
		return &github.Amd64
	}
	return &github.Arm64
}
