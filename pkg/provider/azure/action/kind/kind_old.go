package kind

import (
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	kindApi "github.com/redhat-developer/mapt/pkg/targets/service/kind"
)

func Create(mCtxArgs *mc.ContextArgs, args *kindApi.KindArgs) (*kindApi.KindResults, error) {

	kindRequest :=
		&azureLinux.LinuxArgs{
			Prefix:         args.Prefix,
			ComputeRequest: args.ComputeRequest,
			Spot:           args.Spot,
			Version:        args.Version,
			Arch:           args.Arch,
			OSType:         data.Fedora,
			GetUserdata:    kindApi.GetAsUserdata,
			// As RHEL now is set with cloud init this is the ReadinessCommand to check
			ReadinessCommand: command.CommandCloudInitWait}
	err := azureLinux.Create(mCtxArgs, kindRequest)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func Destroy(ctx *mc.ContextArgs) error {
	return azureLinux.Destroy(ctx)
}
