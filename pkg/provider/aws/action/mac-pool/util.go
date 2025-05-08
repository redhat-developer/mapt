package macpool

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"time"

	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	macUtil "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/util"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// This function will fill information about machines in the pool
// depending on their state and age full fill the struct to easily
// manage them
func getPool(poolName, arch, osVersion string) (*pool, error) {
	// Get machines in the pool
	poolID := &macHost.PoolID{
		PoolName:  poolName,
		Arch:      arch,
		OSVersion: osVersion,
	}
	var p pool
	var err error
	p.machines, err = macHost.GetPoolDedicatedHostsInformation(poolID)
	if err != nil {
		return nil, err
	}
	// non-locked
	p.currentOfferedMachines = util.ArrayFilter(p.machines,
		func(h *mac.HostInformation) bool {
			isLocked, err := macUtil.IsMachineLocked(h)
			if err != nil {
				logging.Errorf("error checking locking for machine %s", *h.Host.AssetId)
				return false
			}
			return !isLocked
		})
	// non-locked + older than 24 hours
	macAgeDestroyRequeriemnt := time.Now().UTC().
		Add(-24 * time.Hour)
	p.destroyableMachines = util.ArrayFilter(p.currentOfferedMachines,
		func(h *mac.HostInformation) bool {
			return h.Host.AllocationTime.UTC().Before(macAgeDestroyRequeriemnt)
		})
	p.name = poolName
	return &p, nil
}

func ticket() (*string, error) {
	timestamp := time.Now().UnixNano() / 1e6
	var randomBytes [2]byte
	_, err := rand.Read(randomBytes[:])
	if err != nil {
		return nil, err
	}
	randomPart := binary.BigEndian.Uint16(randomBytes[:])

	ticket := fmt.Sprintf("%d%04d", timestamp, randomPart)
	return &ticket, nil
}

func writeTicket(ticket *string) error {
	return util.If(len(maptContext.GetResultsOutputPath()) > 0,
		os.WriteFile(path.Join(maptContext.GetResultsOutputPath(), "ticket"), []byte(*ticket), 0600),
		nil)
}
