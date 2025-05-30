package host

import (
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
)

type MacDedicatedHostRequestArgs struct {
	// Allow orquestrate
	Prefix string

	Architecture string
	// Previously it supported check multi region for capacity due to pool approach
	// for the time being this will be fixed
	// FixedLocation bool
	// House keeper requires extra info for setup network and security for managed machines
	VPCID    *string
	Region   *string
	AZID     *string
	SubnetID *string
	SSHSGID  *string
}

type PoolID struct {
	PoolName  string
	Arch      string
	OSVersion string
}

func (p *PoolID) AsTags() map[string]string {
	return map[string]string{
		macConstants.TagKeyArch:      p.Arch,
		macConstants.TagKeyOSVersion: p.OSVersion,
		macConstants.TagKeyPoolName:  p.PoolName,
	}
}

type PoolMacDedicatedHostRequestArgs struct {
	BackedURL        string
	MacDedicatedHost *MacDedicatedHostRequestArgs
	PoolID           *PoolID
}

var (
	awsArchIDbyArch = map[string]string{
		"x86": "x86_64_mac",
		"m1":  "arm64_mac",
		"m2":  "arm64_mac"}
)
