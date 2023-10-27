package mac

const (
	stackDedicatedHost = "stackDedicatedHost"
	stackMacMachine    = "stackMacMachine"

	awsMacMachineID = "amm"

	cidrVN       = "10.0.0.0/16"
	cidrPublicSN = "10.0.2.0/24"
	cidrIntraSN  = "10.0.101.0/24"

	outputHost            = "ammHost"
	outputUsername        = "ammUsername"
	outputUserPassword    = "ammUserPassword"
	outputUserPrivateKey  = "ammUserPrivatekey"
	outputDedicatedHostID = "ammDedicatedHostID"
	outputDedicatedHostAZ = "ammDedicatedHostAZ"
	outputRegion          = "ammRegion"
	// outputAdminUsername     = "ammAdminUsername"
	// outputAdminUserPassword = "ammAdminUserPassword"
	outputBastionUserPrivateKey = "bastionUserPrivatekey"
	outputBastionUsername       = "bastionUsername"
	outputBastionHost           = "bastionHost"

	amiRegex = "amzn-ec2-macos-%s*"
	amiOwner = "628277914472"

	vncDefaultPort  int    = 5900
	diskSize        int    = 200
	defaultUsername string = "ec2-user"
	defaultSSHPort  int    = 22

	// https://www.pulumi.com/docs/intro/concepts/resources/options/customtimeouts/
	remoteTimeout string = "40m"
)

var macTypesByArch = map[string]string{
	"x86": "mac1.metal",
	"m1":  "mac2.metal",
	"m2":  "mac2-m2pro.metal"}

var awsArchIDbyArch = map[string]string{
	"x86": "x86_64_mac",
	"m1":  "arm64_mac",
	"m2":  "arm64_mac"}
