package mac

import (
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	StackDedicatedHost = "stackDedicatedHost"
	StackMacMachine    = "stackMacMachine"
)

type HostInformation struct {
	Arch        *string
	OSVersion   *string
	BackedURL   *string
	Prefix      *string
	ProjectName *string
	RunID       *string
	Region      *string
	Host        *ec2Types.Host
}

var (
	TypesByArch = map[string]string{
		"x86": "mac1.metal",
		"m1":  "mac2.metal",
		"m2":  "mac2-m2pro.metal"}

	AWSArchIDbyArch = map[string]string{
		"x86": "x86_64_mac",
		"m1":  "arm64_mac",
		"m2":  "arm64_mac"}
)
