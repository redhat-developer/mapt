package hosts

import (
	"fmt"

	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureWindows "github.com/redhat-developer/mapt/pkg/provider/azure/action/windows"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdWindows     = "windows"
	cmdWindowsDesc = "windows operations"

	paramWindowsVersion     = "windows-version"
	paramWindowsVersionDesc = "Major version for windows desktop 10 or 11"
	defaultWindowsVersion   = "11"
	paramFeature            = "windows-featurepack"
	paramFeatureDesc        = "windows feature pack"
	defaultFeature          = "23h2-pro"
	paramAdminUsername      = "admin-username"
	paramAdminUsernameDesc  = "username for admin user. Only rdp accessible within generated password"
	defaultAdminUsername    = "rhqpadmin"

	paramProfile     = "profile"
	paramProfileDesc = "comma seperated list of profiles to apply on the target machine. Profiles available: crc"
)

func GetWindowsDesktopCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdWindows,
		Short: cmdWindowsDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCreateWindowsDesktop(), getDestroyWindowsDesktop())
	return c
}

func getCreateWindowsDesktop() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
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

			// ParseEvictionRate
			var spotToleranceValue = spotAzure.DefaultEvictionRate
			if viper.IsSet(paramSpotTolerance) {
				var ok bool
				spotToleranceValue, ok = spotAzure.ParseEvictionRate(
					viper.GetString(paramSpotTolerance))
				if !ok {
					return fmt.Errorf("%s is not a valid spot tolerance value", viper.GetString(azparams.ParamSpotTolerance))
				}
			}

			// Initialize gh actions runner if needed
			if viper.IsSet(params.InstallGHActionsRunner) {
				err := ghactions.InitGHRunnerArgs(viper.GetString(params.GHActionsRunnerToken),
					viper.GetString(params.GHActionsRunnerName),
					viper.GetString(params.GHActionsRunnerRepo))
				if err != nil {
					logging.Error(err)
				}
			}

			instanceRequest := &instancetypes.AzureInstanceRequest{
				CPUs:       viper.GetInt32(params.CPUs),
				MemoryGib:  viper.GetInt32(params.Memory),
				Arch:       instancetypes.Amd64,
				NestedVirt: viper.GetBool(params.NestedVirt),
			}

			if err := azureWindows.Create(
				&azureWindows.WindowsRequest{
					Prefix:               viper.GetString(params.ProjectName),
					Location:             viper.GetString(paramLocation),
					VMSizes:              viper.GetStringSlice(paramVMSize),
					InstaceTypeRequest:   instanceRequest,
					Version:              viper.GetString(paramWindowsVersion),
					Feature:              viper.GetString(paramFeature),
					Username:             viper.GetString(paramUsername),
					AdminUsername:        viper.GetString(paramAdminUsername),
					Profiles:             viper.GetStringSlice(paramProfile),
					SetupGHActionsRunner: viper.IsSet(params.InstallGHActionsRunner),
					Spot:                 viper.IsSet(azparams.ParamSpot),
					SpotTolerance:        spotToleranceValue}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramLocation, "", defaultLocation, paramLocationDesc)
	flagSet.StringSliceP(paramVMSize, "", []string{}, paramVMSizeDesc)
	flagSet.StringP(paramWindowsVersion, "", defaultWindowsVersion, paramWindowsVersionDesc)
	flagSet.StringP(paramFeature, "", defaultFeature, paramFeatureDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.StringP(paramAdminUsername, "", defaultAdminUsername, paramAdminUsernameDesc)
	flagSet.StringSliceP(paramProfile, "", []string{}, paramProfileDesc)
	flagSet.Bool(azparams.ParamSpot, false, azparams.ParamSpotDesc)
	flagSet.StringP(azparams.ParamSpotTolerance, "", azparams.DefaultSpotTolerance, azparams.ParamSpotToleranceDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroyWindowsDesktop() *cobra.Command {
	return &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
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
			if err := azureWindows.Destroy(); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
