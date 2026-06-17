package data

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	v "github.com/IBM-Cloud/power-go-client/clients/instance"
	ps "github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type SystemPoolRequirements struct {
	CloudInstanceId string
	Memory          float64
	Processors      float64
	PreferredType   string
}

type SystemPoolResult struct {
	SelectedType string
	IsPreferred  bool
}

func SelectSystemType(mCtx *mc.Context, args *SystemPoolRequirements) (*SystemPoolResult, error) {
	client, err := piSystemPoolsClient(mCtx, args.CloudInstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to create system pools client: %w", err)
	}
	pools, err := client.GetSystemPools()
	if err != nil {
		return nil, fmt.Errorf("failed to query system pools: %w", err)
	}
	if len(pools) == 0 {
		return nil, fmt.Errorf("no system pools available in workspace %s", args.CloudInstanceId)
	}

	reqCores := args.Processors
	reqMemory := int64(math.Ceil(args.Memory))

	logPoolSummary(pools)

	if pool, ok := pools[args.PreferredType]; ok {
		if poolHasCapacity(&pool, reqCores, reqMemory) {
			logging.Infof("system type %s has sufficient capacity (requested: %.1f cores, %d GiB memory)",
				args.PreferredType, reqCores, reqMemory)
			return &SystemPoolResult{SelectedType: args.PreferredType, IsPreferred: true}, nil
		}
		logging.Warnf("requested system type %s has insufficient capacity; searching for alternatives...", args.PreferredType)
	} else {
		logging.Warnf("requested system type %s not found in workspace; searching for alternatives...", args.PreferredType)
	}

	type candidate struct {
		name     string
		headroom int64
	}
	var candidates []candidate
	for name, pool := range pools {
		if name == args.PreferredType {
			continue
		}
		if poolHasCapacity(&pool, reqCores, reqMemory) {
			candidates = append(candidates, candidate{
				name:     name,
				headroom: poolAvailableMemory(&pool) - reqMemory,
			})
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf(
			"no system pool has sufficient capacity for %.1f cores and %d GiB memory in workspace %s\n%s",
			reqCores, reqMemory, args.CloudInstanceId, poolCapacitySummary(pools))
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].headroom > candidates[j].headroom
	})

	selected := candidates[0].name
	logging.Infof("auto-selected system type %s (requested %s was unavailable)", selected, args.PreferredType)
	return &SystemPoolResult{SelectedType: selected, IsPreferred: false}, nil
}

func poolHasCapacity(pool *models.SystemPool, cores float64, memory int64) bool {
	if pool.MaxCoresAvailable != nil {
		if pool.MaxCoresAvailable.Cores != nil && *pool.MaxCoresAvailable.Cores >= cores &&
			pool.MaxCoresAvailable.Memory != nil && *pool.MaxCoresAvailable.Memory >= memory {
			return true
		}
	}
	if pool.MaxMemoryAvailable != nil {
		if pool.MaxMemoryAvailable.Cores != nil && *pool.MaxMemoryAvailable.Cores >= cores &&
			pool.MaxMemoryAvailable.Memory != nil && *pool.MaxMemoryAvailable.Memory >= memory {
			return true
		}
	}
	if pool.MaxAvailable != nil {
		return pool.MaxAvailable.Cores != nil && *pool.MaxAvailable.Cores >= cores &&
			pool.MaxAvailable.Memory != nil && *pool.MaxAvailable.Memory >= memory
	}
	return false
}

func poolAvailableMemory(pool *models.SystemPool) int64 {
	if pool.MaxAvailable != nil && pool.MaxAvailable.Memory != nil {
		return *pool.MaxAvailable.Memory
	}
	return 0
}

func logPoolSummary(pools models.SystemPools) {
	for name, pool := range pools {
		cores := float64(0)
		mem := int64(0)
		if pool.MaxAvailable != nil {
			if pool.MaxAvailable.Cores != nil {
				cores = *pool.MaxAvailable.Cores
			}
			if pool.MaxAvailable.Memory != nil {
				mem = *pool.MaxAvailable.Memory
			}
		}
		logging.Infof("  system pool %-8s: max available %.1f cores, %d GiB memory", name, cores, mem)
	}
}

func poolCapacitySummary(pools models.SystemPools) string {
	var lines []string
	for name, pool := range pools {
		cores := float64(0)
		mem := int64(0)
		if pool.MaxAvailable != nil {
			if pool.MaxAvailable.Cores != nil {
				cores = *pool.MaxAvailable.Cores
			}
			if pool.MaxAvailable.Memory != nil {
				mem = *pool.MaxAvailable.Memory
			}
		}
		lines = append(lines, fmt.Sprintf("  %-8s: %.1f cores, %d GiB memory available", name, cores, mem))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func piSystemPoolsClient(mCtx *mc.Context, cloudInstanceId string) (*v.IBMPISystemPoolClient, error) {
	options := &ps.IBMPIOptions{
		Authenticator: &core.IamAuthenticator{
			ApiKey: os.Getenv(icConstants.EnvIBMCloudAPIKey),
		},
		UserAccount: os.Getenv(icConstants.EnvIBMCloudAccount),
		Zone:        os.Getenv("IC_ZONE"),
		URL:         powerURL(os.Getenv("IC_REGION")),
		Debug:       mCtx.Debug(),
	}
	session, err := ps.NewIBMPISession(options)
	if err != nil {
		return nil, err
	}
	return v.NewIBMPISystemPoolClient(mCtx.Context(), session, cloudInstanceId), nil
}
