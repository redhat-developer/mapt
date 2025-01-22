package mac

import (
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	macUtil "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/util"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// Request could be interpreted as a general way to create / get a machine
//
// Some project will request a mac machine
// based on tags it will check if there is any existing mac machine (based on labels + arch + MaxPoolSize)
//
// if dh machine exists and max pool size has been reached:
//
// if it exists it will check if it is locked
// if no locked it will replace based on version TODO think how this would afferct airgap / proxy mechanism
// it locked it will (wait or return an error)
//
// if dh does not exist or max pool size has not been reched
// create the machine
//
//	...
func Request(ctx *maptContext.ContextArgs, r *MacRequestArgs) error {
	// Create mapt Context
	maptContext.Init(ctx)

	// Get list of dedicated host ordered by allocation time
	his, err := macHost.GetMatchingHostsInformation(r.Architecture)
	if err != nil {
		return err
	}
	// If no machines we will create one
	if len(his) == 0 {
		return create(r, nil)
	}
	// Pick the most suited to be offered to the requester
	// and replcae (create fresh env)
	// If for whatever reason the mac has no been created
	// stack does nt exist pick will require create not replace
	hi, err := macUtil.PickHost(r.Prefix, his)
	if err != nil {
		if hi == nil {
			return err
		}
		return create(r, hi)
	}
	mr := r.fillMacRequest()
	err = mr.ReplaceUserAccess(hi)
	if err != nil {
		return err
	}
	// We update the runID on the dedicated host
	return tag.Update(maptContext.TagKeyRunID,
		maptContext.RunID(),
		*hi.Region,
		*hi.Host.HostId)
}

// Release will use dedicated host ID as identifier
//
// It will get the info for the dedicated host
// get backedURL (tag on the dh)
// get projectName (tag on the dh)
// load machine stack based on those params
// run release update on it
func Release(ctx *maptContext.ContextArgs, hostID string) error {
	// Get host as context will be fullfilled with info coming from the tags on the host
	host, err := data.GetDedicatedHost(hostID)
	if err != nil {
		return err
	}
	hi := macHost.GetHostInformation(*host)
	// Create mapt Context
	ctx.ProjectName = *hi.ProjectName
	ctx.BackedURL = *hi.BackedURL
	maptContext.Init(ctx)
	// replace machine
	return macMachine.ReplaceMachine(hi)
}

// Initial scenario consider 1 machine
// If we request destroy mac machine it will look for any machine
// and check if it is locked if not locked it will destroy it
func Destroy(ctx *maptContext.ContextArgs, hostID string) error {
	host, err := data.GetDedicatedHost(hostID)
	if err != nil {
		return err
	}
	hi := macHost.GetHostInformation(*host)
	// Create mapt Context
	ctx.ProjectName = *hi.ProjectName
	ctx.BackedURL = *hi.BackedURL
	maptContext.Init(ctx)
	// Dedicated host is not on a valid state to be deleted
	// With same backedURL check if machine is locked
	machineLocked, err := macUtil.IsMachineLocked(hi)
	if err != nil {
		return err
	}
	if !machineLocked {
		if err := aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: mac.StackMacMachine,
			Region:    *hi.Region,
			BackedURL: *hi.BackedURL,
		}); err != nil {
			return err
		}
		return aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: mac.StackDedicatedHost,
			// TODO check if needed to add region for backedURL
			Region:    *hi.Region,
			BackedURL: *hi.BackedURL,
		})
	}
	logging.Debug("nothing to be destroyed")
	return nil
}

// Create function will create the dedicated host
// it will add several tags to it
//
// * backedURL will be used to create the mac machine stack
// * arch will be used to match new requests for specific arch
// * origin fixed mapt value
// * instaceTagName id for the mapt execution
// * (customs) any tag passed as --tags param on the create operation
//
// It will also create a mac machine based on the arch and version setup
// and will set a lock on it

func create(r *MacRequestArgs, dh *mac.HostInformation) (err error) {
	if dh == nil {
		hr := r.fillHostRequest()
		// Get data required for create a dh
		dh, err = macHost.CreateDedicatedHost(hr)
		if err != nil {
			return err
		}
	}
	mr := r.fillMacRequest()
	// Setup the topology and install the mac machine
	if !r.Airgap {

		return mr.CreateAndLockMacMachine(dh)
	}
	return mr.CreateAirgapMacMachine(dh)
}

func (r *MacRequestArgs) fillHostRequest() *macHost.MacDedicatedHostRequestArgs {
	return &macHost.MacDedicatedHostRequestArgs{
		Prefix:        r.Prefix,
		Architecture:  r.Architecture,
		FixedLocation: r.FixedLocation,
	}
}

func (r *MacRequestArgs) fillMacRequest() *macMachine.Request {
	return &macMachine.Request{
		Prefix:               r.Prefix,
		Architecture:         r.Architecture,
		Version:              r.Version,
		SetupGHActionsRunner: r.SetupGHActionsRunner,
		Airgap:               r.Airgap,
	}
}
