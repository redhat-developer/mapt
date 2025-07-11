package params

import (
	"fmt"

	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
	"github.com/spf13/viper"
)

const (
	ParamLocation                = "location"
	ParamLocationDesc            = "location for created resources in case spot flag (if available) is not passed"
	DefaultLocation              = "West US"
	ParamVMSize                  = "vmsize"
	ParamVMSizeDesc              = "size for the VM"
	DefaultVMSize                = "Standard_D8as_v5"
	ParamSpot                    = "spot"
	ParamSpotDesc                = "if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)"
	ParamSpotTolerance           = "spot-eviction-tolerance"
	ParamSpotToleranceDesc       = "if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest"
	DefaultSpotTolerance         = "lowest"
	ParamSpotExcludedRegions     = "spot-excluded-regions"
	ParamSpotExcludedRegionsDesc = "this params allows to pass a comma separated list of regions to avoid when searching for best spot option"
)

func SpotTolerance() (*spotTypes.Tolerance, error) {
	// ParseEvictionRate
	spotToleranceValue := spotTypes.DefaultTolerance
	if viper.IsSet(ParamSpotTolerance) {
		var ok bool
		spotToleranceValue, ok = spotTypes.ParseTolerance(
			viper.GetString(ParamSpotTolerance))
		if !ok {
			return nil, fmt.Errorf("%s is not a valid spot tolerance value",
				viper.GetString(ParamSpotTolerance))
		}
	}
	return &spotToleranceValue, nil
}
