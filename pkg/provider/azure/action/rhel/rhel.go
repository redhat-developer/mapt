package rhel

import (
	"fmt"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	rhelApi "github.com/redhat-developer/mapt/pkg/target/host/rhel"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type RhelArgs struct {
	Prefix         string
	Location       string
	Arch           string
	ComputeRequest *cr.ComputeRequestArgs
	Version        string
	SubsUsername   string
	SubsUserpass   string
	ProfileSNC     bool
	Username       string
	Spot           *spotTypes.SpotArgs
}

func imageRef(version, arch string) *data.ImageReference {
	if arch == "arm64" {
		return &data.ImageReference{
			Publisher: "rhel-arm64",
			Offer:     "RHEL",
			Sku: fmt.Sprintf(
				"%s-arm64",
				strings.ReplaceAll(version, ".", "_"))}
	}
	return &data.ImageReference{
		Publisher: "RedHat",
		Offer:     "RHEL",
		Sku: fmt.Sprintf(
			"%s-gen2",
			strings.ReplaceAll(version, ".", ""))}
}

func Create(mCtxArgs *maptContext.ContextArgs, r *RhelArgs) (err error) {
	logging.Debug("Creating RHEL Server")
	rhelCloudConfig := &rhelApi.CloudConfigArgs{
		SNCProfile:   r.ProfileSNC,
		SubsUsername: r.SubsUsername,
		SubsPassword: r.SubsUserpass,
		Username:     r.Username}
	azureLinuxRequest :=
		&azureLinux.LinuxArgs{
			Prefix:         r.Prefix,
			Location:       r.Location,
			ComputeRequest: r.ComputeRequest,
			Spot:           r.Spot,
			ImageRef:       imageRef(r.Version, r.Arch),
			Version:        r.Version,
			Arch:           r.Arch,
			OSType:         data.RHEL,
			Username:       r.Username,
			// As RHEL now is set with cloud init this is the ReadinessCommand to check
			CloudConfigAsUserData: rhelCloudConfig,
			ReadinessCommand:      command.CommandCloudInitWait}
	return azureLinux.Create(mCtxArgs, azureLinuxRequest)
}

func Destroy(mCtxArgs *maptContext.ContextArgs) error {
	return azureLinux.Destroy(mCtxArgs)
}
