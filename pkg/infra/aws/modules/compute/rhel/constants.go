package rhel

const (
	// https://access.redhat.com/solutions/15356
	defaultAMIPattern string = "RHEL-%s*-x86_64-*"

	// for each ami we should know the default user, otherwise need to manage users from userdata
	defaultAMIUser string = "ec2-user"
	// arm
	// defaultInstanceType string = "c6g.metal"
	defaultInstanceType string = "c5n.metal"
	// defaultBlockDurationMinutes int    = 120

	VERSION_8 = "8"

	// bastionDefaultDeviceType   string = "gp2"
	// bastionDefaultDeviceSize   int    = 10

	OutputPrivateIP  string = "rhelPrivateIP"
	OutputUsername   string = "rhelUsername"
	OutputPrivateKey string = "rhelPrivateKey"
)

// var instanceType []string = []string{"c6g.metal"}
