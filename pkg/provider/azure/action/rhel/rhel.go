package rhel

import (
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	cloudConfigRHEL "github.com/redhat-developer/mapt/pkg/provider/util/cloud-config/rhel"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type Request struct {
	Prefix              string
	Location            string
	VMSizes             []string
	Arch                string
	InstanceRequest     instancetypes.InstanceRequest
	Version             string
	SubsUsername        string
	SubsUserpass        string
	ProfileSNC          bool
	Username            string
	Spot                bool
	SpotTolerance       spotAzure.EvictionRate
	SpotExcludedRegions []string
	Timeout             string
}

func Create(ctx *maptContext.ContextArgs, r *Request) (err error) {
	if len(r.VMSizes) == 0 {
		vmSizes, err := r.InstanceRequest.GetMachineTypes()
		if err != nil {
			logging.Debugf("Unable to fetch desired instance type: %v", err)
		}
		if len(vmSizes) > 0 {
			r.VMSizes = append(r.VMSizes, vmSizes...)
		}
	}
	logging.Debug("Creating RHEL Server")
	rhelCloudConfig := &cloudConfigRHEL.RequestArgs{
		SNCProfile:   r.ProfileSNC,
		SubsUsername: r.SubsUsername,
		SubsPassword: r.SubsUserpass,
		Username:     r.Username}
	azureLinuxRequest :=
		&azureLinux.LinuxRequest{
			Prefix:          r.Prefix,
			Location:        r.Location,
			VMSizes:         r.VMSizes,
			InstanceRequest: r.InstanceRequest,
			Version:         r.Version,
			Arch:            r.Arch,
			OSType:          data.RHEL,
			Username:        r.Username,
			Spot:            r.Spot,
			SpotTolerance:   r.SpotTolerance,
			GetUserdata:     rhelCloudConfig.GetAsUserdata,
			// As RHEL now is set with cloud init this is the ReadinessCommand to check
			ReadinessCommand: command.CommandCloudInitWait}
	return azureLinux.Create(ctx, azureLinuxRequest)
}

func Destroy(ctx *maptContext.ContextArgs) error {
	return azureLinux.Destroy(ctx)
}
