package bastion

const (
	// bastionDefaultAMI          string = "Amazon Linux 2 AMI (HVM)"
	bastionDefaultAMI string = "amzn-ami-hvm-*-x86_64-ebs"
	// for each ami we should know the default user, otherwise need to manage users from userdata
	bastionDefaultAMIUser      string = "ec2-user"
	bastionDefaultInstanceType string = "t2.small"
	// bastionDefaultDeviceType   string = "gp2"
	// bastionDefaultDeviceSize   int    = 10

	OutputPublicIP   string = "bastionPublicIP"
	OutputUsername   string = "bastionUsername"
	OutputPrivateKey string = "bastionPrivateKey"
)
