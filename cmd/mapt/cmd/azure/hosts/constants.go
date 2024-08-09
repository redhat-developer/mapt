package hosts

const (
	paramLocation          = "location"
	paramLocationDesc      = "If spot is passed location will be calculated based on spot results. Otherwise localtion will be used to create resources."
	defaultLocation        = "West US"
	paramVMSize            = "vmsize"
	paramVMSizeDesc        = "size for the VM. Type requires to allow nested virtualization"
	defaultVMSize          = "Standard_D8as_v5"
	paramUsername          = "username"
	paramUsernameDesc      = "username for general user. SSH accessible + rdp with generated password"
	defaultUsername        = "rhqp"
	paramSpot              = "spot"
	paramSpotDesc          = "if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)"
	paramSpotTolerance     = "spot-eviction-tolerance"
	paramSpotToleranceDesc = "if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest"
	defaultSpotTolerance   = "lowest"
)
