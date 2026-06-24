package params

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/integrations/gitlab"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
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
	diskSize            string = "disk-size"
	diskSizeDesc        string = "Disk size in GB for the cloud instance"
	diskSizeDefault     int    = 200

	CreateCmdName  string = "create"
	DestroyCmdName string = "destroy"

	ghActionsRunnerToken    string = "ghactions-runner-token"
	ghActionsRunnerRepo     string = "ghactions-runner-repo"
	ghActionsRunnerLabels   string = "ghactions-runner-labels"
	ghActionsRunnerImageRepo string = "ghactions-runner-image-repo"
	// TODO: once the RHEL script is merged to https://github.com/IBM/action-runner-image-pz,
	// switch default from deekay2310 fork to IBM upstream.
	ghActionsRunnerImageRepoDefault string = "https://github.com/deekay2310/action-runner-image-pz.git"
	GHActionsRunnerImageRepoDesc    string = "Git clone URL for the action-runner-image-pz repository, used to build the GitHub Actions runner from source on ppc64le/s390x (no official binaries exist for these architectures)"

	cirrusPWToken      string = "it-cirrus-pw-token"
	cirrusPWTokenDesc  string = "Add mapt target as a cirrus persistent worker. The value will hold a valid token to be used by cirrus cli to join the project."
	cirrusPWLabels     string = "it-cirrus-pw-labels"
	cirrusPWLabelsDesc string = "additional labels to use on the persistent worker (--it-cirrus-pw-labels key1=value1,key2=value2)"

	glRunnerToken         string = "glrunner-token"
	glRunnerTokenDesc     string = "GitLab token with create_runner scope (personal access token, group/project access token, or service account token)"
	glRunnerProjectID     string = "glrunner-project-id"
	glRunnerProjectIDDesc string = "GitLab project ID for project runner registration"
	glRunnerGroupID       string = "glrunner-group-id"
	glRunnerGroupIDDesc   string = "GitLab group ID for group runner registration (alternative to --glrunner-project-id)"
	glRunnerURL           string = "glrunner-url"
	glRunnerURLDesc       string = "GitLab instance URL (e.g., https://gitlab.com, https://gitlab.example.com)"
	glRunnerURLDefault    string = "https://gitlab.com"
	glRunnerTags          string = "glrunner-tags"
	glRunnerTagsDesc      string = "List of tags separated by comma to be added to the self-hosted runner"
	glRunnerUnsecure string = "glrunner-unsecure"
	glRunnerUnsecureDesc string = "when set, the runner service runs as the default OS user instead of a dedicated system account; by default a locked-down gitlab-runner system user is created"

	GlRunnerConcurrent             string = "glrunner-concurrent"
	GlRunnerConcurrentDesc         string = "maximum number of jobs the runner executes concurrently"
	GlRunnerConcurrentPowerDefault int    = 2
	GlRunnerConcurrentS390xDefault int    = 3

	//RHEL
	SubsUsername              string = "rh-subscription-username"
	SubsUsernameDesc          string = "username to register the subscription"
	SubsUserpass              string = "rh-subscription-password"
	SubsUserpassDesc          string = "password to register the subscription"
	ProfileSNC                string = "snc"
	ProfileSNCDesc            string = "if this flag is set the RHEL will be setup with SNC profile. Setting up all requirements to run https://github.com/crc-org/snc"
	RhelVersion               string = "version"
	RhelVersionDesc           string = "version for the RHEL OS"
	RhelVersionDefault        string = "9.4"
	RhelAIVersion             string = "version"
	RhelAIVersionDesc         string = "version for the RHELAI OS"
	RhelAIVersionDefault      string = "3.0.0"
	RhelAIAccelerator         string = "accelerator"
	RhelAIAccelearatorDesc    string = "accelerator type. Valid types: cuda and rocm"
	RhelAIAccelearatorDefault string = "cuda"
	RhelAICustomImage         string = "custom-image"
	RhelAICustomImageDesc     string = "custom image name to spin RHEL AI OS (AMI name for AWS, image name for Azure)"

	// Serverless
	Timeout        string = "timeout"
	TimeoutDesc    string = "if timeout is set a serverless destroy actions will be set on the time according to the timeout. The Timeout value is a duration conforming to Go ParseDuration format."
	Serverless     string = "serverless"
	ServerlessDesc string = "if serverless is set the command will be executed as a serverless action."

	// Destroy
	ForceDestroy     string = "force-destroy"
	ForceDestroyDesc string = "if force-destroy is set the command will destroy even if there is a lock."
	KeepState        string = "keep-state"
	KeepStateDesc    string = "keep Pulumi state files in backend storage after successful destroy (by default, state files are removed)"

	// IBM Cloud
	SubnetID              string = "subnet-id"
	SubnetIDDesc          string = "ID of an existing VPC subnet to deploy the instance into"
	WorkspaceID           string = "workspace-id"
	WorkspaceIDDesc       string = "ID of an existing Power VS workspace (cloud instance)"
	PIPrivateSubnetID     string = "pi-private-subnet-id"
	PIPrivateSubnetIDDesc string = "ID of an existing Power VS private subnet to attach the instance to"
	VPCPublicSubnetID     string = "vpc-public-subnet-id"
	VPCPublicSubnetIDDesc string = "ID of an existing VPC subnet (with public gateway, connected to Transit Gateway) for the SSH bastion"

	// IBM Power instance sizing
	PIMemory             string  = "pi-memory"
	PIMemoryDesc         string  = "PowerVS instance memory in GB"
	PIMemoryDefault      float64 = 96.0
	PIProcessors         string  = "pi-processors"
	PIProcessorsDesc     string  = "PowerVS instance processor count (shared cores)"
	PIProcessorsDefault  float64 = 24.0
	PIProcType           string  = "pi-proc-type"
	PIProcTypeDesc       string  = "PowerVS processor type (shared, dedicated, capped)"
	PIProcTypeDefault    string  = "shared"
	PISysType            string  = "pi-sys-type"
	PISysTypeDesc        string  = "preferred PowerVS system type (e.g. e1080, s1022, s1122); if unset, auto-discovered from zone"
	PISysTypeDefault     string  = ""
	PIStorageType        string  = "pi-storage-type"
	PIStorageTypeDesc    string  = "PowerVS storage tier for instance and data volume (tier1, tier3)"
	PIStorageTypeDefault string  = "tier1"
	PIDiskSize           string  = "pi-disk-size"
	PIDiskSizeDesc       string  = "data volume size in GB attached to the PowerVS instance"
	PIDiskSizeDefault    int     = 300

	// IBM Z instance sizing
	IZProfile         string = "iz-profile"
	IZProfileDesc     string = "IBM Z VPC instance profile name"
	IZProfileDefault  string = "mz2-16x128"
	IZDiskSize        string = "iz-disk-size"
	IZDiskSizeDesc    string = "boot volume size in GB for the IBM Z instance (10-250 for general-purpose profile)"
	IZDiskSizeDefault int    = 250

	OtelAppCode       string = "otel-app-code"
	OtelAppCodeDesc   string = "OpenTelemetry appcode identifier (e.g. MAPT-001); when set together with --otel-auth-token, installs the otelcol-contrib filelog collector on the instance"
	OtelAuthToken     string = "otel-auth-token"
	OtelAuthTokenDesc string = "OpenTelemetry authentication token (UUID) used to authenticate against the OTLP endpoint"
	OtelEndpoint      string = "otel-endpoint"
	OtelEndpointDesc  string = "OTLP HTTP endpoint to export logs to"
	OtelIndex          string = "otel-index"
	OtelIndexDesc      string = "Splunk index name for log routing (e.g. rh_linux)"
	OtelExtraAttrs     string = "otel-extra-attrs"
	OtelExtraAttrsDesc string = "Additional resource attributes to attach to all otelcol log records (key=value pairs)"

	// Kind
	KindCmd                   = "kind"
	KindCmdDesc               = "Manage a Kind cluster. This is not intended for production use"
	KindK8SVersion            = "version"
	KindK8SVersionDesc        = "version for k8s offered through Kind."
	KindK8SVersionDefault     = "v1.34"
	KindExtraPortMappings     = "extra-port-mappings"
	KindExtraPortMappingsDesc = "Additional port mappings for the Kind cluster. Value should be a JSON array of objects with containerPort, hostPort, and protocol properties. Example: '[{\"containerPort\": 8080, \"hostPort\": 8080, \"protocol\": \"TCP\"}]'"

	// Network
	ServiceEndpoints = "service-endpoints"

	// Spot
	spot                         = "spot"
	spotDesc                     = "if spot is set the spot prices across all regions will be checked and machine will be started on best spot option (price / eviction)"
	spotTolerance                = "spot-eviction-tolerance"
	spotToleranceDesc            = "if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest"
	spotToleranceDefault         = "lowest"
	spotPriceIncreaseRate        = "spot-increase-rate"
	spotPriceIncreaseRateDesc    = "Percentage to be added on top of the current calculated spot price to increase chances to get the machine"
	spotPriceIncreaseRateDefault = 30
	spotExcludedHostedZones      = "spot-excluded-regions"
	spotExcludedHostedZonesDesc  = "Comma-separated list of zone IDs to exclude from spot selection"
)

func AddNetworkFlags(fs *pflag.FlagSet, desc string) {
	fs.StringSliceP(ServiceEndpoints, "", []string{}, desc)
}

func NetworkServiceEndpoints() []string {
	return viper.GetStringSlice(ServiceEndpoints)
}

func AddSpotFlags(fs *pflag.FlagSet) {
	fs.Bool(spot, false, spotDesc)
	fs.StringP(spotTolerance, "", spotToleranceDefault, spotToleranceDesc)
	fs.IntP(spotPriceIncreaseRate, "", spotPriceIncreaseRateDefault, spotPriceIncreaseRateDesc)
	fs.StringSliceP(spotExcludedHostedZones, "", []string{}, spotExcludedHostedZonesDesc)
}

func SpotArgs() *spotTypes.SpotArgs {
	if viper.IsSet(spot) {
		sa := &spotTypes.SpotArgs{
			Spot:                  viper.IsSet(spot),
			IncreaseRate:          viper.GetInt(spotPriceIncreaseRate),
			ExcludedHostingPlaces: viper.GetStringSlice(spotExcludedHostedZones),
		}
		if t, b := spotTypes.ParseTolerance(viper.GetString(spotTolerance)); b {
			sa.Tolerance = t
		}
		return sa
	}
	return nil
}

func AddComputeRequestFlags(fs *pflag.FlagSet) {
	fs.Int32P(cpus, "", 8, cpusDesc)
	fs.Int32P(gpus, "", 0, gpusDesc)
	fs.StringP(gpuManufacturer, "", "", gpuManufacturerDesc)
	fs.Int32P(memory, "", 64, memoryDesc)
	fs.BoolP(nestedVirt, "", false, nestedVirtDesc)
	fs.StringSliceP(computeSizes, "", []string{}, computeSizesDesc)
	fs.IntP(diskSize, "", diskSizeDefault, diskSizeDesc)
}

func ComputeRequestArgs() *cr.ComputeRequestArgs {
	cra := &cr.ComputeRequestArgs{
		CPUs:            viper.GetInt32(cpus),
		GPUs:            viper.GetInt32(gpus),
		GPUManufacturer: viper.GetString(gpuManufacturer),
		MemoryGib:       viper.GetInt32(memory),
		Arch: util.If(viper.GetString(LinuxArch) == "arm64",
			cr.Arm64, cr.Amd64),
		NestedVirt:   viper.GetBool(ProfileSNC) || viper.GetBool(nestedVirt),
		ComputeSizes: viper.GetStringSlice(computeSizes),
	}
	if viper.IsSet(diskSize) {
		ds := viper.GetInt(diskSize)
		cra.DiskSize = &ds
	}
	return cra
}

func AddCommonFlags(fs *pflag.FlagSet) {
	fs.StringP(ProjectName, "", "", ProjectNameDesc)
	fs.StringP(BackedURL, "", "", BackedURLDesc)
}

func AddDebugFlags(fs *pflag.FlagSet) {
	fs.Bool(Debug, false, DebugDesc)
	fs.Uint(DebugLevel, DebugLevelDefault, DebugLevelDesc)
}

func AddGHActionsFlags(fs *pflag.FlagSet) {
	fs.StringP(ghActionsRunnerToken, "", "", GHActionsRunnerTokenDesc)
	fs.StringP(ghActionsRunnerRepo, "", "", GHActionsRunnerRepoDesc)
	fs.StringSlice(ghActionsRunnerLabels, nil, GHActionsRunnerLabelsDesc)
	fs.StringP(ghActionsRunnerImageRepo, "", ghActionsRunnerImageRepoDefault, GHActionsRunnerImageRepoDesc)
}

func GithubRunnerArgs() *github.GithubRunnerArgs {
	token := viper.GetString(ghActionsRunnerToken)
	repoURL := viper.GetString(ghActionsRunnerRepo)
	pat := os.Getenv("GITHUB_TOKEN")

	if token == "" && pat == "" {
		return nil
	}

	if repoURL == "" {
		logging.Error("--ghactions-runner-repo is required for GitHub Actions runner setup")
		return nil
	}

	if token == "" {
		logging.Info("no --ghactions-runner-token provided, auto-generating from GITHUB_TOKEN")
		var err error
		token, err = github.GenerateRegistrationToken(pat, repoURL)
		if err != nil {
			logging.Errorf("failed to auto-generate runner registration token: %v", err)
			return nil
		}
		logging.Info("runner registration token generated successfully")
	}

	imageRepo := viper.GetString(ghActionsRunnerImageRepo)
	if imageRepo != "" {
		if err := validateRunnerImageRepo(imageRepo); err != nil {
			logging.Errorf("invalid --ghactions-runner-image-repo: %v", err)
			return nil
		}
		if imageRepo != ghActionsRunnerImageRepoDefault {
			logging.Infof("using custom runner image repo: %s", imageRepo)
		} else {
			logging.Debugf("using temporary fork %s; will switch to IBM upstream once RHEL script is merged", imageRepo)
		}
	}
	return &github.GithubRunnerArgs{
		Token:           token,
		RepoURL:         repoURL,
		Labels:          viper.GetStringSlice(ghActionsRunnerLabels),
		Platform:        &github.Linux,
		Arch:            linuxArchAsGithubActionsArch(viper.GetString(LinuxArch)),
		RunnerImageRepo: imageRepo,
	}
}

func validateRunnerImageRepo(repo string) error {
	if !strings.HasPrefix(repo, "https://") {
		return fmt.Errorf("only HTTPS URLs are allowed, got: %s", repo)
	}
	return nil
}

func AddCirrusFlags(fs *pflag.FlagSet) {
	fs.StringP(cirrusPWToken, "", "", cirrusPWTokenDesc)
	fs.StringToStringP(cirrusPWLabels, "", nil, cirrusPWLabelsDesc)
}

func AddGitLabRunnerFlags(fs *pflag.FlagSet) {
	fs.StringP(glRunnerToken, "", "", glRunnerTokenDesc)
	fs.StringP(glRunnerProjectID, "", "", glRunnerProjectIDDesc)
	fs.StringP(glRunnerGroupID, "", "", glRunnerGroupIDDesc)
	fs.StringP(glRunnerURL, "", glRunnerURLDefault, glRunnerURLDesc)
	fs.StringSlice(glRunnerTags, nil, glRunnerTagsDesc)
	fs.Bool(glRunnerUnsecure, false, glRunnerUnsecureDesc)
}


func CirrusPersistentWorkerArgs() *cirrus.PersistentWorkerArgs {
	if viper.IsSet(cirrusPWToken) {
		return &cirrus.PersistentWorkerArgs{
			Token:    viper.GetString(cirrusPWToken),
			Labels:   viper.GetStringMapString(cirrusPWLabels),
			Platform: &cirrus.Linux,
			Arch: linuxArchAsCirrusArch(
				viper.GetString(LinuxArch)),
		}
	}
	return nil
}

func GitLabRunnerArgs(arch *gitlab.Arch) *gitlab.GitLabRunnerArgs {
	if viper.IsSet(glRunnerToken) {
		if viper.IsSet(glRunnerProjectID) && viper.IsSet(glRunnerGroupID) {
			logging.Error("--glrunner-project-id and --glrunner-group-id are mutually exclusive; ignoring GitLab runner configuration")
			return nil
		}
		return &gitlab.GitLabRunnerArgs{
			GitLabToken: viper.GetString(glRunnerToken),
			ProjectID:   viper.GetString(glRunnerProjectID),
			GroupID:     viper.GetString(glRunnerGroupID),
			URL:         viper.GetString(glRunnerURL),
			Tags:        viper.GetStringSlice(glRunnerTags),
			Platform:    &gitlab.Linux,
			Arch:        arch,
			Unsecure:    viper.GetBool(glRunnerUnsecure),
			Concurrent:  viper.GetInt(GlRunnerConcurrent),
		}
	}
	return nil
}

func LinuxGitLabArch() *gitlab.Arch {
	return linuxArchAsGitLabArch(viper.GetString(LinuxArch))
}

func linuxArchAsCirrusArch(arch string) *cirrus.Arch {
	switch arch {
	case "x86_64":
		return &cirrus.Amd64
	}
	return &cirrus.Arm64
}

func linuxArchAsGithubActionsArch(arch string) *github.Arch {
	switch arch {
	case "x86_64":
		return &github.Amd64
	case "ppc64le":
		return &github.Ppc64le
	case "s390x":
		return &github.S390x
	}
	return &github.Arm64
}

func MACArchAsCirrusArch(arch string) *cirrus.Arch {
	switch arch {
	case "x86":
		return &cirrus.Amd64
	}
	return &cirrus.Arm64
}

func linuxArchAsGitLabArch(arch string) *gitlab.Arch {
	switch arch {
	case "x86_64":
		return &gitlab.Amd64
	}
	return &gitlab.Arm64
}

func MACArchAsGitLabArch(arch string) *gitlab.Arch {
	switch arch {
	case "x86":
		return &gitlab.Amd64
	}
	return &gitlab.Arm64
}

