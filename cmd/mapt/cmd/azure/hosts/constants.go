package hosts

const (
	paramVMSize           = "vmsize"
	paramVMSizeDesc       = "set specific size for the VM and ignore any CPUs, Memory and Arch parameters set. Type requires to allow nested virtualization"
	paramUsername         = "username"
	paramUsernameDesc     = "username for general user. SSH accessible + rdp with generated password"
	defaultUsername       = "rhqp"
	paramLinuxVersion     = "version"
	paramLinuxVersionDesc = "linux version. Version should be formated as X.Y (Major.minor)"
	defaultUbuntuVersion  = "24.04"
	defaultRHELVersion    = "9.4"
	defaultFedoraVersion  = "42"
)
