package rhel

import (
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	cloudConfigRHEL "github.com/redhat-developer/mapt/pkg/provider/util/cloud-config/rhel"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type RhelArgs struct {
	Prefix              string
	Location            string
	Arch                string
	ComputeRequest      *cr.ComputeRequestArgs
	Version             string
	SubsUsername        string
	SubsUserpass        string
	ProfileSNC          bool
	Username            string
	Spot                bool
	SpotTolerance       spot.Tolerance
	SpotExcludedRegions []string
}

func Create(ctx *maptContext.ContextArgs, r *RhelArgs) (err error) {
	logging.Debug("Creating RHEL Server")
	rhelCloudConfig := &cloudConfigRHEL.RequestArgs{
		SNCProfile:   r.ProfileSNC,
		SubsUsername: r.SubsUsername,
		SubsPassword: r.SubsUserpass,
		Username:     r.Username}
	azureLinuxRequest :=
		&azureLinux.LinuxArgs{
			Prefix:         r.Prefix,
			Location:       r.Location,
			ComputeRequest: r.ComputeRequest,
			Version:        r.Version,
			Arch:           r.Arch,
			OSType:         data.RHEL,
			Username:       r.Username,
			Spot:           r.Spot,
			SpotTolerance:  r.SpotTolerance,
			GetUserdata:    rhelCloudConfig.GetAsUserdata,
			// As RHEL now is set with cloud init this is the ReadinessCommand to check
			ReadinessCommand: command.CommandCloudInitWait}
	return azureLinux.Create(ctx, azureLinuxRequest)
}

func Destroy(ctx *maptContext.ContextArgs) error {
	return azureLinux.Destroy(ctx)
}
