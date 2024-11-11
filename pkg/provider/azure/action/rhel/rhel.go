package rhel

import (
	"fmt"

	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	targetRHEL "github.com/redhat-developer/mapt/pkg/targets/rhel"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type Request struct {
	Prefix          string
	Location        string
	VMSizes         []string
	Arch            string
	InstanceRequest instancetypes.InstanceRequest
	Version         string
	SubsUsername    string
	SubsUserpass    string
	ProfileSNC      bool
	Username        string
	Spot            bool
	SpotTolerance   spotAzure.EvictionRate
	// setup as github actions runner
	SetupGHActionsRunner bool
}

func Create(r *Request) (err error) {
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
	userDataB64, err := targetRHEL.GetUserdata(r.ProfileSNC,
		r.SubsUsername, r.SubsUserpass, r.Username,
		r.SetupGHActionsRunner)
	if err != nil {
		return fmt.Errorf("error creating RHEL Server on Azure: %v", err)
	}
	azureLinuxRequest :=
		&azureLinux.LinuxRequest{
			Prefix:          r.Prefix,
			Location:        r.Location,
			VMSizes:         r.VMSizes,
			InstanceRequest: r.InstanceRequest,
			Version:         r.Version,
			Arch:            r.Arch,
			OSType:          azureLinux.RHEL,
			Username:        r.Username,
			Spot:            r.Spot,
			SpotTolerance:   r.SpotTolerance,
			Userdata:        userDataB64,
			// As RHEL now is set with cloud init this is the ReadinessCommand to check
			ReadinessCommand: command.CommandCloudInitWait}
	return azureLinux.Create(azureLinuxRequest)
}

func Destroy() error {
	return azureLinux.Destroy()
}
