package host

import (
	"github.com/go-playground/validator/v10"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

const (
	// mapt internal ID for the component: nac dedicated host
	awsMacHostID = "amh"

	tagKeyPrefix    = "prefix"
	tagKeyBackedURL = "backedURL"
	tagKeyArch      = "arch"
	// tags added when dedicated host is part of a pool
	tagKeyOSVersion = "osVersion"
	tagKeyPoolName  = "poolName"

	outputDedicatedHostID = "ammDedicatedHostID"
	outputDedicatedHostAZ = "ammDedicatedHostAZ"
	outputRegion          = "ammRegion"
)

type MacDedicatedHostRequestArgs struct {
	// Allow orquestrate
	Prefix string

	Architecture  string
	FixedLocation bool
}

type PoolID struct {
	PoolName  string
	Arch      string
	OSVersion string
}

func (p *PoolID) asTags() map[string]string {
	return map[string]string{
		tagKeyArch:      p.Arch,
		tagKeyOSVersion: p.OSVersion,
		tagKeyPoolName:  p.PoolName,
	}
}

type PoolMacDedicatedHostRequestArgs struct {
	BackedURL        string
	MacDedicatedHost *MacDedicatedHostRequestArgs
	PoolID           *PoolID
}

type dedicatedHostArgs struct {
	mCtx             *mc.Context `validate:"required"`
	prefix           string
	arch             string
	region           *string
	availabilityZone *string
	tags             map[string]string
}

func (r *dedicatedHostArgs) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	err := v.Var(r.mCtx, "required")
	if err != nil {
		return err
	}
	return v.Struct(r)
}

var (
	awsArchIDbyArch = map[string]string{
		"x86": "x86_64_mac",
		"m1":  "arm64_mac",
		"m2":  "arm64_mac"}
)
