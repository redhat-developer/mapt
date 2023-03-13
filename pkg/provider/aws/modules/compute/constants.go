package compute

const (
	OutputHost        string = "Host"
	OutputUsername    string = "Username"
	OutputPrivateKey  string = "PrivateKey"
	OutputPasswordKey string = "Password"

	DefaultRootBlockDeviceName string = "/dev/sda1"
	DefaultRootBlockDeviceSize int    = 100

	// Delay health check due to baremetal + userdata otherwise it will kill hosts consntantly
	// Probably move this to compute asset as each can have different value depending on userdata
	defaultHealthCheckGracePeriod int = 1200
)
