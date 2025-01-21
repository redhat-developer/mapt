package constants

import "github.com/redhat-developer/mapt/pkg/provider/aws/action/mac"

const (
	MACRequestCmd      = "request"
	MACRRequestCmdDesc = "request mac machine"
	MACReleaseCmd      = "release"
	MACReleaseCmdDesc  = "release mac machine"

	MACArch              string = "arch"
	MACArchDesc          string = "MAC architecture allowed values x86, m1, m2"
	MACArchDefault       string = mac.DefaultArch
	MACOSVersion         string = "version"
	MACOSVersionDesc     string = "MACos operating system vestion 11, 12 on x86 and m1/m2; 13, 14 on all archs"
	MACOSVersionDefault  string = mac.DefaultOSVersion
	MACFixedLocation     string = "fixed-location"
	MACFixedLocationDesc string = "if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)"
	MACDHID              string = "dedicated-host-id"
	MACDHIDDesc          string = "id for the dedicated host"
)
