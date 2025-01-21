package hosts

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
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
	cmdMac     = "mac"
	cmdMacDesc = "manage mac instances"
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
		Use:   awsParams.MACRequestCmd,
		Short: awsParams.MACRequestCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Initialize context
			maptContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel),
				false)

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
					Architecture:         viper.GetString(awsParams.MACArch),
					Version:              viper.GetString(awsParams.MACOSVersion),
					FixedLocation:        viper.IsSet(awsParams.MACFixedLocation),
					SetupGHActionsRunner: viper.GetBool(params.InstallGHActionsRunner),
					Airgap:               viper.IsSet(airgap)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACRequestCmd, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.Bool(awsParams.MACFixedLocation, false, awsParams.MACFixedLocationDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

// Required dedicatedHostID as mandatory
func getMacRelease() *cobra.Command {
	c := &cobra.Command{
		Use:   awsParams.MACReleaseCmd,
		Short: awsParams.MACReleaseCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Run create
			if err := mac.Release(
				"main",
				viper.GetString(awsParams.MACDHID),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACReleaseCmd, pflag.ExitOnError)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(awsParams.MACDHID, "", "", awsParams.MACDHIDDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(awsParams.MACDHID)
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
				viper.GetString(awsParams.MACDHID),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.StringP(awsParams.MACDHID, "", "", awsParams.MACDHIDDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(awsParams.MACDHID)
	if err != nil {
		logging.Error(err)
	}
	return c
}
