package rhelai

import (
	"fmt"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"
	"github.com/redhat-developer/mapt/pkg/provider/util/command"
	apiRHELAI "github.com/redhat-developer/mapt/pkg/targets/host/rhelai"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	imageOwnerSubscriptionId = "02db6bd4-035c-4074-b699-468f3d914744"
	// $1 accelerator $2 version
	imageNameRegex = "rhel-ai-%s-azure-%s"
	// $1 subscriptionId $2 rgName
	imageIdRegex = "/subscriptions/%s/resourceGroups/aipcc-productization/providers/Microsoft.Compute/galleries/%s/images/%s/versions/1.0.0"

	username = "azureuser"
)

func imageId(accelerator, version string) string {
	iName := fmt.Sprintf(imageNameRegex, accelerator, version)
	gName := strings.ReplaceAll(iName, "-", "_")
	return fmt.Sprintf(imageIdRegex,
		imageOwnerSubscriptionId,
		gName,
		iName)
}

func Create(mCtxArgs *maptContext.ContextArgs, args *apiRHELAI.RHELAIArgs) (err error) {
	logging.Debug("Creating RHEL Server")
	azureLinuxRequest :=
		&azureLinux.LinuxArgs{
			Prefix: args.Prefix,
			// Location:         args.Location,
			ComputeRequest: args.ComputeRequest,
			Spot:           args.Spot,
			ImageRef: &data.ImageReference{
				SharedImageID: imageId(args.Accelerator, args.Version),
			},
			Username:         username,
			ReadinessCommand: command.CommandPing}
	return azureLinux.Create(mCtxArgs, azureLinuxRequest)
}

func Destroy(mCtxArgs *maptContext.ContextArgs) error {
	return azureLinux.Destroy(mCtxArgs)
}
