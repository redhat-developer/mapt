package hosts

const (
	spot       string = "spot"
	spotDesc   string = "if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)"
	airgap     string = "airgap"
	airgapDesc string = "if this flag is set the host will be created as airgap machine. Access will done through a bastion"

	vmTypes            string = "vm-types"
	vmTypesDescription string = "set an specific set of vm-types and ignore any CPUs, Memory, Arch parameters set. Note vm-type should match requested arch. Also if --spot flag is used set at least 3 types."
)
