package params

import (
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ProjectName                 string = "project-name"
	ProjectNameDesc             string = "project name to identify the instance of the stack"
	BackedURL                   string = "backed-url"
	BackedURLDesc               string = "backed for stack state. (local) file:///path/subpath (s3) s3://existing-bucket, (azure) azblob://existing-blobcontainer. See more https://www.pulumi.com/docs/iac/concepts/state-and-backends/#using-a-self-managed-backend"
	ConnectionDetailsOutput     string = "conn-details-output"
	ConnectionDetailsOutputDesc string = "path to export host connection information (host, username and privateKey)"
	Debug                       string = "debug"
	DebugDesc                   string = "Enable debug traces and set verbosity to max. Typically to get information to troubleshooting an issue."
	DebugLevel                  string = "debug-level"
	DebugLevelDefault           uint   = 3
	DebugLevelDesc              string = "Set the level of verbosity on debug. You can set from minimum 1 to max 9."
	LinuxArch                   string = "arch"
	LinuxArchDesc               string = "architecture for the machine. Allowed x86_64 or arm64"
	LinuxArchDefault            string = "x86_64"
	SupportedHostID             string = "host-id"
	SupportedHostIDDesc         string = "host id from supported hosts list"
	AvailabilityZones           string = "availability-zones"
	AvailabilityZonesDesc       string = "List of comma separated azs to check. If empty all will be searched"
	RHMajorVersion              string = "rh-major-version"
	RHMajorVersionDesc          string = "major version for rhel image 7, 8 or 9"
	RHSubcriptionUsername       string = "rh-subscription-username"
	RHSubcriptionUsernameDesc   string = "username for rhel subcription"
	RHSubcriptionPassword       string = "rh-subscription-password"
	RHSubcriptionPasswordDesc   string = "password for rhel subcription"
	FedoraMajorVersion          string = "fedora-major-version"
	FedoraMajorVersionDesc      string = "major version for fedora image 36, 37, ..."
	MacOSMajorVersion           string = "macos-major-version"
	MacOSMajorVersionDesc       string = "major version for macos image 12, 13, ..."
	AMIIDName                   string = "ami-id"
	AMIIDDesc                   string = "id for the source ami"
	AMINameName                 string = "ami-name"
	AMINameDesc                 string = "name for the source ami"
	AMISourceRegion             string = "ami-region"
	AMISourceRegionDesc         string = "region for the ami to be copied worldwide"
	Tags                        string = "tags"
	TagsDesc                    string = "tags to add on each resource (--tags name1=value1,name2=value2)"
	GHActionsRunnerTokenDesc    string = "Token needed for registering the Github Actions Runner token"
	GHActionsRunnerRepoDesc     string = "Full URL of the repository where the Github Actions Runner should be registered"
	GHActionsRunnerLabelsDesc   string = "List of labels separated by comma to be added to the self-hosted runner"

	// Compute request
	memory              string = "memory"
	memoryDesc          string = "Amount of RAM for the cloud instance in GiB"
	cpus                string = "cpus"
	cpusDesc            string = "Number of CPUs for the cloud instance"
	gpus                string = "gpus"
	gpusDesc            string = "Number of GPUs for the cloud instance"
	gpuManufacturer     string = "gpu-manufacturer"
	gpuManufacturerDesc string = "Manufacturer company name for GPU. (i.e. NVIDIA)"
	nestedVirt          string = "nested-virt"
	nestedVirtDesc      string = "Use cloud instance that has nested virtualization support"
	computeSizes        string = "compute-sizes"
	computeSizesDesc    string = "Comma seperated list of sizes for the machines to be requested. If set this takes precedence over compute by args"

	CreateCmdName  string = "create"
	DestroyCmdName string = "destroy"

	GHActionsRunnerToken  string = "ghactions-runner-token"
	GHActionsRunnerRepo   string = "ghactions-runner-repo"
	GHActionsRunnerLabels string = "ghactions-runner-labels"

	CirrusPWToken      string = "it-cirrus-pw-token"
	CirrusPWTokenDesc  string = "Add mapt target as a cirrus persistent worker. The value will hold a valid token to be used by cirrus cli to join the project."
	CirrusPWLabels     string = "it-cirrus-pw-labels"
	CirrusPWLabelsDesc string = "additional labels to use on the persistent worker (--it-cirrus-pw-labels key1=value1,key2=value2)"

	//RHEL
	SubsUsername         string = "rh-subscription-username"
	SubsUsernameDesc     string = "username to register the subscription"
	SubsUserpass         string = "rh-subscription-password"
	SubsUserpassDesc     string = "password to register the subscription"
	ProfileSNC           string = "snc"
	ProfileSNCDesc       string = "if this flag is set the RHEL will be setup with SNC profile. Setting up all requirements to run https://github.com/crc-org/snc"
	RhelVersion          string = "version"
	RhelVersionDesc      string = "version for the RHEL OS"
	RhelVersionDefault   string = "9.4"
	RhelAIVersion        string = "version"
	RhelAIVersionDesc    string = "version for the RHELAI OS"
	RhelAIVersionDefault string = "1.5.0"

	// Serverless
	Timeout        string = "timeout"
	TimeoutDesc    string = "if timeout is set a serverless destroy actions will be set on the time according to the timeout. The Timeout value is a duration conforming to Go ParseDuration format."
	Serverless     string = "serverless"
	ServerlessDesc string = "if serverless is set the command will be executed as a serverless action."

	// Desytoy
	ForceDestroy     string = "force-destroy"
	ForceDestroyDesc string = "if force-destroy is set the command will destroy even if there is a lock."

	// Kind
	KindCmd            = "kind"
	KindCmdDesc        = "Manage a Kind cluster. This is not intended for production use"
	KindK8SVersion     = "version"
	KindK8SVersionDesc = "version for k8s offered through Kind."

	// Spot
	SpotPriceIncreaseRate        = "spot-increase-rate"
	SpotPriceIncreaseRateDesc    = "Percentage to be added on top of the current calculated spot price to increase chances to get the machine"
	SpotPriceIncreaseRateDefault = 20
)

func GetGHActionsFlagset() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(GHActionsRunnerToken, "", "", GHActionsRunnerTokenDesc)
	flagSet.StringP(GHActionsRunnerRepo, "", "", GHActionsRunnerRepoDesc)
	flagSet.StringSlice(GHActionsRunnerLabels, nil, GHActionsRunnerLabelsDesc)
	return flagSet
}

func GetCpusAndMemoryFlagset() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(CreateCmdName, pflag.ExitOnError)
	flagSet.Int32P(cpus, "", 8, cpusDesc)
	flagSet.Int32P(gpus, "", 0, gpusDesc)
	flagSet.StringP(gpuManufacturer, "", "", gpuManufacturerDesc)
	flagSet.Int32P(memory, "", 64, memoryDesc)
	flagSet.BoolP(nestedVirt, "", false, nestedVirtDesc)
	flagSet.StringSliceP(computeSizes, "", []string{}, computeSizesDesc)
	return flagSet
}

func GetComputeRequest() *cr.ComputeRequestArgs {
	return &cr.ComputeRequestArgs{
		CPUs:            viper.GetInt32(cpus),
		GPUs:            viper.GetInt32(gpus),
		GPUManufacturer: viper.GetString(gpuManufacturer),
		MemoryGib:       viper.GetInt32(memory),
		Arch: util.If(viper.GetString(LinuxArch) == "arm64",
			cr.Arm64, cr.Amd64),
		NestedVirt:   viper.GetBool(ProfileSNC) || viper.GetBool(nestedVirt),
		ComputeSizes: viper.GetStringSlice(computeSizes),
	}
}

func AddCommonFlags(fs *pflag.FlagSet) {
	fs.StringP(ProjectName, "", "", ProjectNameDesc)
	fs.StringP(BackedURL, "", "", BackedURLDesc)
}

func AddDebugFlags(fs *pflag.FlagSet) {
	fs.Bool(Debug, false, DebugDesc)
	fs.Uint(DebugLevel, DebugLevelDefault, DebugLevelDesc)
}

func AddCirrusFlags(fs *pflag.FlagSet) {
	fs.StringP(CirrusPWToken, "", "", CirrusPWTokenDesc)
	fs.StringToStringP(CirrusPWLabels, "", nil, CirrusPWLabelsDesc)
}

func LinuxArchAsCirrusArch(arch string) *cirrus.Arch {
	switch arch {
	case "x86_64":
		return &cirrus.Amd64
	}
	return &cirrus.Arm64
}

func LinuxArchAsGithubActionsArch(arch string) *github.Arch {
	switch arch {
	case "x86_64":
		return &github.Amd64
	}
	return &github.Arm64
}
