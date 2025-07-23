package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
	macMachine "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/machine"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	StackDedicatedHost = "stackDedicatedHost"
	StackMacMachine    = "stackMacMachine"

	outputLock = "ammLock"
)

// We will get a list of hosts from the pool ordered by allocation time
// We will apply several rules on them to pick the right one
// - TODO Remove those with allocation time > 24 h as they may destroyed
// - if none left use them again
// - if more available pick in order the first without lock
func PickHost(prefix string, his []*mac.HostInformation) (*mac.HostInformation, error) {
	for _, h := range his {
		isLocked, err := IsMachineLocked(h)
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

func IsMachineLocked(h *mac.HostInformation) (bool, error) {
	s, err := manager.CheckStack(manager.Stack{
		StackName:   fmt.Sprintf("%s-%s", StackMacMachine, *h.ProjectName),
		ProjectName: *h.ProjectName,
		BackedURL:   *h.BackedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				awsConstants.CONFIG_AWS_REGION: *h.Region}),
	})
	if err != nil {
		return false, err
	}
	outputs, err := manager.GetOutputs(s)
	if err != nil {
		return false, err
	}
	if value, exists := outputs[fmt.Sprintf("%s-%s", *h.Prefix, outputLock)]; exists {
		return value.Value.(bool), nil
	}
	return false, errors.New("stack outputs does not contain value for lock so we assume is not locked")
}

// Release will use dedicated host ID as identifier
//
// It will get the info for the dedicated host
// get backedURL (tag on the dh)
// get projectName (tag on the dh)
// load machine stack based on those params
// run release update on it
func Release(ctx *mc.ContextArgs, hostID string) error {
	// Get host as context will be fullfilled with info coming from the tags on the host
	host, err := data.GetDedicatedHost(hostID)
	if err != nil {
		return err
	}
	hi := macHost.GetHostInformation(*host)
	// Create mapt Context
	ctx.ProjectName = *hi.ProjectName
	ctx.BackedURL = *hi.BackedURL
	mCtx, err := mc.Init(ctx, aws.Provider())
	if err != nil {
		return err
	}
	// replace machine
	return macMachine.ReplaceMachine(mCtx, hi)
}
