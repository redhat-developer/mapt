package hosts

const (
	paramLocation          = "location"
	paramLocationDesc      = "If spot is passed location will be calculated based on spot results. Otherwise localtion will be used to create resources."
	defaultLocation        = "West US"
	paramVMSize            = "vmsize"
	paramVMSizeDesc        = "set specific size for the VM and ignore any CPUs, Memory and Arch parameters set. Type requires to allow nested virtualization"
	paramUsername          = "username"
	paramUsernameDesc      = "username for general user. SSH accessible + rdp with generated password"
	defaultUsername        = "rhqp"
	paramSpot              = "spot"
	paramSpotDesc          = "if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)"
	paramSpotTolerance     = "spot-eviction-tolerance"
	paramSpotToleranceDesc = "if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest"
	defaultSpotTolerance   = "lowest"
	paramLinuxVersion      = "version"
	paramLinuxVersionDesc  = "linux version. Version should be formated as X.Y (Major.minor)"
	defaultUbuntuVersion   = "24.04"
	defaultRHELVersion     = "9.4"
	defaultFedoraVersion   = "40.0"
)
