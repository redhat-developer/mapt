package constants

const (
	ParamLocation          = "location"
	ParamLocationDesc      = "location for created resources in case spot flag (if available) is not passed"
	DefaultLocation        = "West US"
	ParamVMSize            = "vmsize"
	ParamVMSizeDesc        = "size for the VM"
	DefaultVMSize          = "Standard_D8as_v5"
	ParamSpot              = "spot"
	ParamSpotDesc          = "if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)"
	ParamSpotTolerance     = "spot-eviction-tolerance"
	ParamSpotToleranceDesc = "if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest"
	DefaultSpotTolerance   = "lowest"
)
