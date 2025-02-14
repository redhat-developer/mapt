package machine

import (
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/network"
)

type Request struct {
	// Prefix for the resources related to mac
	// this is relevant in case of an orchestration with multiple
	// macs on the same stack
	Prefix           string
	Region           *string
	AvailabilityZone *string
	Version          string
	Architecture     string
	// setup as github actions runner
	SetupGHActionsRunner bool
	Airgap               bool
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
	// For airgap scenario there is an orchestation of
	// a phase with connectivity on the machine (allowing bootstraping)
	// a pahase with connectivyt off where the subnet for the target lost the nat gateway
	airgapPhaseConnectivity network.Connectivity
	// dh linkage
	dedicatedHost *mac.HostInformation
	// operation control params
	isRequestOperation   bool
	lock                 bool
	sshConnectionTimeout string
	// values used to increase security on request operations
	// currentPrivateKey pulumi.StringPtrInput
	// currentPassword   pulumi.StringPtrInput
	currentPrivateKey string
	currentPassword   string
}

const (
	awsMacMachineID = "amm"

	customResourceTypeLock = "rh:qe:aws:mac:lock"
	customResourceTypeKey  = "rh:mapt:aws:mac:key"

	outputLock              = "ammLock"
	outputHost              = "ammHost"
	outputUsername          = "ammUsername"
	outputUserPassword      = "ammUserPassword"
	outputMachinePrivateKey = "ammMachinePrivatekey"
	outputUserPrivateKey    = "ammUserPrivatekey"
	outputDedicatedHostID   = "ammDedicatedHostID"
	outputDedicatedHostAZ   = "ammDedicatedHostAZ"
	outputRegion            = "ammRegion"

	amiRegex         = "amzn-ec2-macos-%s*"
	DefaultArch      = "m2"
	DefaultOSVersion = "15"

	vncDefaultPort  int    = 5900
	diskSize        int    = 100
	blockDeviceType string = "gp3"
	defaultUsername string = "ec2-user"
	defaultSSHPort  int    = 22

	// https://www.pulumi.com/docs/intro/concepts/resources/options/customtimeouts/
	defaultTimeout string = "30m"
	releaseTimeout string = "25m"
	requestTimeout string = "5m"
)

var awsArchIDbyArch = map[string]string{
	"x86": "x86_64_mac",
	"m1":  "arm64_mac",
	"m2":  "arm64_mac"}

func isAWSArchID(a string) bool { return a == "x86_64_mac" || a == "arm64_mac" }
