package macpool

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	macConstants "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/constants"
	macHost "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/host"
)

// this is a business identificator to assing to resources related to the serverless management
func serverlessName(poolName, arch, osVersion, operation string) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		operation,
		poolName,
		arch,
		osVersion)
}

func serverlessTaskARN(poolName, arch, osVersion, operation string) (*string, error) {
	rARNs, err := data.GetResourcesMatchingTags(
		data.ResourceTypeECS,
		serverlessTags(
			poolName,
			arch,
			osVersion,
			operation))
	if err != nil {
		return nil, err
	}
	if len(rARNs) > 1 {
		return nil, fmt.Errorf(
			"should be only one task spec matching tags. Found %s",
			strings.Join(rARNs, ","))
	}
	return &rARNs[0], nil

}

// Return the map of tags wich should identify unique
// resquest operation spec for a pool
func serverlessTags(poolName, arch, osVersion, operation string) (m map[string]string) {
	poolID := macHost.PoolID{
		PoolName:  poolName,
		Arch:      arch,
		OSVersion: osVersion,
	}
	m = poolID.AsTags()
	m[macConstants.TagKeyPoolOperationName] = operation
	return
}
