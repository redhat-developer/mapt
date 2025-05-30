package host

import (
	"maps"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	dedicatedhost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/ec2/compute/dedicated-host"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilMaps "github.com/redhat-developer/mapt/pkg/util/maps"
)

func CreatePoolDedicatedHost(args *PoolMacDedicatedHostRequestArgs) (dhi *mac.HostInformation, err error) {
	tags := map[string]string{
		macConstants.TagKeyBackedURL: args.BackedURL,
		macConstants.TagKeyPrefix:    args.MacDedicatedHost.Prefix,
		maptContext.TagKeyRunID:      maptContext.RunID(),
	}
	maps.Copy(tags, args.PoolID.AsTags())
	return createDedicatedHost(args.MacDedicatedHost, args.BackedURL, tags, false)
}

func DestroyPoolDedicatedHost(prefix *string) error {
	return dedicatedhost.Destroy(prefix)
}

// this creates the stack for the dedicated host
func CreateDedicatedHost(args *MacDedicatedHostRequestArgs) (dhi *mac.HostInformation, err error) {
	backedURL := getBackedURL()
	tags := map[string]string{
		macConstants.TagKeyBackedURL: backedURL,
		macConstants.TagKeyPrefix:    args.Prefix,
		macConstants.TagKeyArch:      args.Architecture,
		maptContext.TagKeyRunID:      maptContext.RunID(),
		macConstants.TagKeyTicket:    "",
	}
	return createDedicatedHost(args, backedURL, tags, true)
}

func createDedicatedHost(args *MacDedicatedHostRequestArgs,
	backedURL string,
	tags map[string]string,
	exportOutputs bool) (dhi *mac.HostInformation, err error) {
	dhTags := utilMaps.Append(maptContext.GetTags(), tags)
	it := mac.TypesByArch[args.Architecture]
	dHArgs := dedicatedhost.DedicatedHostArgs{
		Prefix:       &args.Prefix,
		InstanceType: &it,
		Tags:         dhTags,
	}
	dHArgs.Region, err = getRegion(args.Architecture, true)
	if err != nil {
		return nil, err
	}
	dh, err := dedicatedhost.Create(&backedURL, exportOutputs, &dHArgs)
	if err != nil {
		return nil, err
	}
	logging.Debugf("mac dedicated host with host id %s has been created successfully", *dh.HostId)
	host, err := data.GetDedicatedHost(*dh.HostId)
	if err != nil {
		return nil, err
	}
	i := GetHostInformation(*host)
	dhi = i
	return
}
