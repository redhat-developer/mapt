package constants

import (
	"github.com/spf13/pflag"
)

const (
	ProjectName                 string = "project-name"
	ProjectNameDesc             string = "project name to identify the instance of the stack"
	BackedURL                   string = "backed-url"
	BackedURLDesc               string = "backed for stack state. Can be a local path with format file:///path/subpath or s3 s3://existing-bucket"
	ConnectionDetailsOutput     string = "conn-details-output"
	ConnectionDetailsOutputDesc string = "path to export host connection information (host, username and privateKey)"
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
	InstallGHActionsRunnerDesc  string = "Install and setup Github Actions runner in the instance"
	GHActionsRunnerTokenDesc    string = "Token needed for registering the Github Actions Runner token"
	GHActionsRunnerNameDesc     string = "Name for the Github Actions Runner"
	GHActionsRunnerRepoDesc     string = "Full URL of the repository where the Github Actions Runner should be registered"
	Memory                      string = "memory"
	MemoryDesc                  string = "Amount of RAM for the cloud instance in GiB"
	CPUs                        string = "cpus"
	CPUsDesc                    string = "Number of CPUs for the cloud instance"
	NestedVirt                  string = "nested-virt"
	NestedVirtDesc              string = "Use cloud instance that has nested virtualization support"

	CreateCmdName  string = "create"
	DestroyCmdName string = "destroy"

	InstallGHActionsRunner string = "install-ghactions-runner"
	GHActionsRunnerToken   string = "ghactions-runner-token"
	GHActionsRunnerName    string = "ghactions-runner-name"
	GHActionsRunnerRepo    string = "ghactions-runner-repo"
)

func GetGHActionsFlagset() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(CreateCmdName, pflag.ExitOnError)
	flagSet.Bool(InstallGHActionsRunner, false, InstallGHActionsRunnerDesc)
	flagSet.StringP(GHActionsRunnerToken, "", "", GHActionsRunnerTokenDesc)
	flagSet.StringP(GHActionsRunnerName, "", "", GHActionsRunnerNameDesc)
	flagSet.StringP(GHActionsRunnerRepo, "", "", GHActionsRunnerRepoDesc)
	return flagSet
}

func GetCpusAndMemoryFlagset() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(CreateCmdName, pflag.ExitOnError)
	flagSet.Int32P(CPUs, "", 8, CPUsDesc)
	flagSet.Int32P(Memory, "", 64, MemoryDesc)
	flagSet.BoolP(NestedVirt, "", false, NestedVirtDesc)
	return flagSet
}
