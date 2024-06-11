package mac

import (
	_ "embed"
	"fmt"
	"strings"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/services/tag"
	"github.com/redhat-developer/mapt/pkg/util/logging"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Request could be interpreted as a general way to create / release
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
func Request(r *MacRequest) error {
	// Get list of dedicated host ordered by allocation time
	his, err := getMatchingHostsInformation(r.Architecture)
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
	hi, err := pickHost(r.Prefix, his)
	if err != nil {
		if hi == nil {
			return err
		}
		return create(r, hi)
	}
	err = r.replaceMachine(hi)
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
func Release(prefix string, hostID string) error {
	host, err := data.GetDedicatedHost(hostID)
	if err != nil {
		return err
	}
	hi := getHostInformation(*host)
	// Set context based on info from dedicated host to be released
	maptContext.InitBase(
		*hi.ProjectName,
		*hi.BackedURL)
	// Set a default request
	r := &MacRequest{
		Prefix:       prefix,
		Architecture: archDefault,
		Version:      osVersionDefault,
	}
	return r.releaseLock(hi)
}

// Initial scenario consider 1 machine
// If we request destroy mac machine it will look for any machine
// and check if it is locked if not locked it will destroy it
func Destroy(prefix, hostID string) error {
	host, err := data.GetDedicatedHost(hostID)
	if err != nil {
		return err
	}
	hi := getHostInformation(*host)
	// Set context based on info from dedicated host to be released
	maptContext.InitBase(
		*hi.ProjectName,
		*hi.BackedURL)
	// Check if dh is available and it has no instance on it
	// otherwise we can not release it
	if hi.Host.State == ec2Types.AllocationStateAvailable &&
		len(host.Instances) == 0 {
		return aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: stackDedicatedHost,
			// TODO check if needed to add region for backedURL
			Region:    *hi.Region,
			BackedURL: *hi.BackedURL,
		})
	}
	// Dedicated host is not on a valid state to be deleted
	// With same backedURL check if machine is locked
	machineLocked, err := isMachineLocked(prefix, hi)
	if err != nil {
		return err
	}
	if !machineLocked {
		return aws.DestroyStack(aws.DestroyStackRequest{
			Stackname: stackMacMachine,
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

func create(r *MacRequest, dh *HostInformation) (err error) {
	if dh == nil {
		dh, err = r.createDedicatedHost()
		if err != nil {
			return err
		}
	}
	// Setup the topology and install the mac machine
	if !r.Airgap {
		return r.createMacMachine(dh)
	}
	return r.createAirgapMacMachine(dh)
}

// We will get a list of hosts from the pool ordered by allocation time
// We will apply several rules on them to pick the right one
// - TODO Remove those with allocation time > 24 h as they may destroyed
// - if none left use them again
// - if more available pick in order the first without lock
func pickHost(prefix string, his []*HostInformation) (*HostInformation, error) {
	for _, h := range his {
		isLocked, err := isMachineLocked(prefix, h)
		if err != nil {
			logging.Errorf("error checking if machine %s is locked", *h.Host.HostId)
			if strings.Contains(err.Error(), "no stack") {
				return h, err
			}
		}
		if !isLocked {
			return h, nil
		}
	}
	return nil, fmt.Errorf("all hosts are locked at the moment")
}
