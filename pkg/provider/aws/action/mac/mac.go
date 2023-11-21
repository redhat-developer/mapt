package mac

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

type MacRequest struct {
	Prefix           string
	Architecture     string
	Version          string
	HostID           string
	OnlyHost         bool
	OnlyMachine      bool
	FixedLocation    bool
	Region           string
	AvailabilityZone string
	Airgap           bool
	// For airgap scenario there is an orchestation of
	// a phase with connectivity on the machine (allowing bootstraping)
	// a pahase with connectivyt off where the subnet for the target lost the nat gateway
	airgapPhaseConnectivity network.Connectivity
}

// this function orchestrate the two stacks related to mac machine
// * the underlaying dedicated host
// * the mac machine
func Create(r *MacRequest) (err error) {
	if len(r.HostID) == 0 {
		// Check if instance type is available on current location
		// region is only needed for dedicated host mac machine got the region from the az
		// of the dedicated host or the request if it creates both at once
		region, err := getRegion(r)
		if err != nil {
			return err
		}
		r.Region = *region
		// No host id means need to create dedicated host
		dhID, dhAZ, err := r.createDedicatedHost()
		if err != nil {
			return err
		}
		r.HostID = *dhID
		r.AvailabilityZone = *dhAZ
	}
	if !r.OnlyHost {
		// if not only host the mac machine will be created
		if !r.Airgap {
			return r.createMacMachine()
		}
		// Airgap scneario requires orchestration
		return r.createAirgapMacMachine()
	}
	return nil
}

// TODO add loop until state with target state and timeout?
func CheckState(hostID string) (*string, error) {
	return data.GetDedicatedHostState(hostID)
}

// Will destroy resources related to machine
func Destroy(r *MacRequest) (err error) {
	var region *string
	if !r.OnlyHost {
		region, err = r.getRegionFromStack(stackMacMachine)
		if err != nil {
			return
		}
		if err = aws.DestroyStackByRegion(*region, stackMacMachine); err != nil {
			return
		}
	}
	if !r.OnlyMachine {
		// We need to get dedicated host region to set on stack
		if region == nil {
			region, err = r.getRegionFromStack(stackDedicatedHost)
			if err != nil {
				return
			}
		}
		return aws.DestroyStackByRegion(*region, stackDedicatedHost)
	}
	return nil
}

// checks if the machine can be created on the current location (machine type is available on the region)
// if it available it returns the region name
// if not offered and machine should be created on the region it will return an error
// if not offered and machine could be created anywhere it will get a region offering the machine and return its name
func getRegion(r *MacRequest) (*string, error) {
	logging.Debugf("checking if %s is offered at %s",
		r.Architecture,
		os.Getenv("AWS_DEFAULT_REGION"))
	isOffered, err := data.IsInstaceTypeOffered(
		macTypesByArch[r.Architecture],
		os.Getenv("AWS_DEFAULT_REGION"))
	if err != nil {
		return nil, err
	}
	if isOffered {
		region := os.Getenv("AWS_DEFAULT_REGION")
		return &region, nil
	}
	if !isOffered && r.FixedLocation {
		return nil, fmt.Errorf("the requested mac %s is not available at the current region %s and the fixed-location flag has been set",
			r.Architecture,
			os.Getenv("AWS_DEFAULT_REGION"))
	}
	// We look for a region offering the type of instance
	return data.LokupRegionOfferingInstanceType(
		macTypesByArch[r.Architecture])
}

// Mac machine can be dinamically moved across regions as it is
// tied to the dedicated host we save the region on the stack to setu[
// the AWS session
func (r *MacRequest) getRegionFromStack(stackName string) (*string, error) {
	stack, err := manager.CheckStack(manager.Stack{
		ProjectName: qenvsContext.GetInstanceName(),
		StackName:   qenvsContext.GetStackInstanceName(stackName),
		BackedURL:   qenvsContext.GetBackedURL()})
	if err != nil {
		return nil, err
	}
	outputs, err := manager.GetOutputs(stack)
	if err != nil {
		return nil, err
	}
	region, ok := outputs[fmt.Sprintf("%s-%s", r.Prefix, outputRegion)].Value.(string)
	if ok {
		return &region, nil
	}
	return nil, fmt.Errorf("%s not found", fmt.Sprintf("%s-%s", r.Prefix, outputRegion))
}
