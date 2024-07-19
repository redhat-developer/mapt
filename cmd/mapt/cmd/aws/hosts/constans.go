package hosts

const (
	spot       string = "spot"
	spotDesc   string = "if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)"
	airgap     string = "airgap"
	airgapDesc string = "if this flag is set the host will be created as airgap machine. Access will done through a bastion"

	linuxArch          string = "arch"
	linuxArchDesc      string = "architecture for the machine. Allowed x86_64 or arm64"
	linuxArchDefault   string = "x86_64"
	vmTypes            string = "vm-types"
	vmTypesDescription string = "set an specific set of vm-types. Note vm-type should match requested arch. Also if --spot flag is used set at least 3 types."
)
