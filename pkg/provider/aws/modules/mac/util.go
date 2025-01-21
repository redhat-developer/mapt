package mac

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

// We will get a list of hosts from the pool ordered by allocation time
// We will apply several rules on them to pick the right one
// - TODO Remove those with allocation time > 24 h as they may destroyed
// - if none left use them again
// - if more available pick in order the first without lock
func PickHost(prefix string, his []*HostInformation) (*HostInformation, error) {
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

func IsMachineLocked(h *HostInformation) (bool, error) {
	s, err := manager.CheckStack(manager.Stack{
		StackName:   maptContext.StackNameByProject(StackMacMachine),
		ProjectName: maptContext.ProjectName(),
		BackedURL:   *h.BackedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: *h.Region}),
	})
	if err != nil {
		return false, err
	}
	outputs, err := manager.GetOutputs(s)
	if err != nil {
		return false, err
	}
	return outputs[fmt.Sprintf("%s-%s", *h.Prefix, outputLock)].Value.(bool), nil
}
