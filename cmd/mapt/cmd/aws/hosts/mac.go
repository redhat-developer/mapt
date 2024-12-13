package hosts

import (
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/mac"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdMac         = "mac"
	cmdMacDesc     = "manage mac instances"
	requestCmd     = "request"
	requestCmdDesc = "request mac machine"
	releaseCmd     = "release"
	releaseCmdDesc = "release mac machine"

	dhID              string = "dedicated-host-id"
	dhIDDesc          string = "id for the dedicated host"
	arch              string = "arch"
	archDesc          string = "mac architecture allowed values x86, m1, m2"
	archDefault       string = mac.DefaultArch
	osVersion         string = "version"
	osVersionDesc     string = "macos operating system vestion 11, 12 on x86 and m1/m2; 13, 14 on all archs"
	osDefault         string = mac.DefaultOSVersion
	fixedLocation     string = "fixed-location"
	fixedLocationDesc string = "if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)"
)

func GetMacCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdMac,
		Short: cmdMacDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getMacRequest(), getMacRelease(), getMacDestroy())
	return c
}

func getMacRequest() *cobra.Command {
	c := &cobra.Command{
		Use:   requestCmd,
		Short: requestCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Initialize context
			maptContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags))

			// Initialize gh actions runner if needed
			if viper.IsSet(params.InstallGHActionsRunner) {
				err := ghactions.InitGHRunnerArgs(viper.GetString(params.GHActionsRunnerToken),
					viper.GetString(params.GHActionsRunnerName),
					viper.GetString(params.GHActionsRunnerRepo),
					viper.GetStringSlice(params.GHActionsRunnerLabels))
				if err != nil {
					logging.Fatal(err)
				}
			}

			// Run create
			if err := mac.Request(
				&mac.MacRequest{
					Prefix:               "main",
					Architecture:         viper.GetString(arch),
					Version:              viper.GetString(osVersion),
					FixedLocation:        viper.IsSet(fixedLocation),
					SetupGHActionsRunner: viper.GetBool(params.InstallGHActionsRunner),
					Airgap:               viper.IsSet(airgap)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(requestCmd, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(arch, "", archDefault, archDesc)
	flagSet.StringP(osVersion, "", osDefault, osVersionDesc)
	flagSet.Bool(fixedLocation, false, fixedLocationDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

// Required dedicatedHostID as mandatory
func getMacRelease() *cobra.Command {
	c := &cobra.Command{
		Use:   releaseCmd,
		Short: releaseCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Run create
			if err := mac.Release(
				"main",
				viper.GetString(dhID)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(releaseCmd, pflag.ExitOnError)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(dhID, "", "", dhIDDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(dhID)
	if err != nil {
		logging.Error(err)
	}
	return c
}

func getMacDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			if err := mac.Destroy(
				"main",
				viper.GetString(dhID)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.StringP(dhID, "", "", dhIDDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(dhID)
	if err != nil {
		logging.Error(err)
	}
	return c
}
