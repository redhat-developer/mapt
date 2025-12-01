package hosts

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/params"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/mac"
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
			return mac.Request(
				&maptContext.ContextArgs{
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					GHRunnerArgs:  params.GithubRunnerArgs(),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&mac.MacRequestArgs{
					Prefix:        "main",
					Architecture:  viper.GetString(awsParams.MACArch),
					Version:       viper.GetString(awsParams.MACOSVersion),
					FixedLocation: viper.IsSet(awsParams.MACFixedLocation),
					Airgap:        viper.IsSet(airgap)})
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACRequestCmd, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.Bool(awsParams.MACFixedLocation, false, awsParams.MACFixedLocationDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	params.AddGHActionsFlags(flagSet)
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
			return mac.Release(
				&maptContext.ContextArgs{
					Debug:      viper.IsSet(params.Debug),
					DebugLevel: viper.GetUint(params.DebugLevel),
				},
				viper.GetString(awsParams.MACDHID))
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACReleaseCmd, pflag.ExitOnError)
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
			return mac.Destroy(
				&maptContext.ContextArgs{
					Debug:        viper.IsSet(params.Debug),
					DebugLevel:   viper.GetUint(params.DebugLevel),
					CleanupState: viper.IsSet(params.CleanupState),
				},
				viper.GetString(awsParams.MACDHID))
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.StringP(awsParams.MACDHID, "", "", awsParams.MACDHIDDesc)
	flagSet.Bool(params.CleanupState, true, params.CleanupStateDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(awsParams.MACDHID)
	if err != nil {
		logging.Error(err)
	}
	return c
}
