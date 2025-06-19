package hosts

import (
	"fmt"

	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureWindows "github.com/redhat-developer/mapt/pkg/provider/azure/action/windows"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
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

			// ParseEvictionRate
			var spotToleranceValue = spotAzure.DefaultEvictionRate
			if viper.IsSet(azparams.ParamSpotTolerance) {
				var ok bool
				spotToleranceValue, ok = spotAzure.ParseEvictionRate(
					viper.GetString(azparams.ParamSpotTolerance))
				if !ok {
					return fmt.Errorf("%s is not a valid spot tolerance value", viper.GetString(azparams.ParamSpotTolerance))
				}
			}

			ctx := &maptContext.ContextArgs{
				ProjectName:   viper.GetString(params.ProjectName),
				BackedURL:     viper.GetString(params.BackedURL),
				ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
				Debug:         viper.IsSet(params.Debug),
				DebugLevel:    viper.GetUint(params.DebugLevel),
				Tags:          viper.GetStringMapString(params.Tags),
			}

			if viper.IsSet(params.CirrusPWToken) {
				ctx.CirrusPWArgs = &cirrus.PersistentWorkerArgs{
					Token:    viper.GetString(params.CirrusPWToken),
					Labels:   viper.GetStringMapString(params.CirrusPWLabels),
					Platform: &cirrus.Windows,
					// Currently we only provide amd64 support for windows
					Arch: &cirrus.Amd64,
				}
			}

			if viper.IsSet(params.GHActionsRunnerToken) {
				ctx.GHRunnerArgs = &github.GithubRunnerArgs{
					Token:    viper.GetString(params.GHActionsRunnerToken),
					RepoURL:  viper.GetString(params.GHActionsRunnerRepo),
					Labels:   viper.GetStringSlice(params.GHActionsRunnerLabels),
					Platform: &github.Windows,
					Arch:     &github.Amd64,
				}
			}

			if err := azureWindows.Create(
				ctx,
				&azureWindows.WindowsRequest{
					Prefix:   viper.GetString(params.ProjectName),
					Location: viper.GetString(paramLocation),
					VMSizes:  viper.GetStringSlice(paramVMSize),
					InstaceTypeRequest: &instancetypes.AzureInstanceRequest{
						CPUs:       viper.GetInt32(params.CPUs),
						MemoryGib:  viper.GetInt32(params.Memory),
						Arch:       instancetypes.Amd64,
						NestedVirt: viper.GetBool(params.NestedVirt),
					},
					Version:             viper.GetString(paramWindowsVersion),
					Feature:             viper.GetString(paramFeature),
					Username:            viper.GetString(paramUsername),
					AdminUsername:       viper.GetString(paramAdminUsername),
					Profiles:            viper.GetStringSlice(paramProfile),
					Spot:                viper.IsSet(azparams.ParamSpot),
					Timeout:             viper.GetString(params.Timeout),
					SpotTolerance:       spotToleranceValue,
					SpotExcludedRegions: viper.GetStringSlice(azparams.ParamSpotExcludedRegions)}); err != nil {
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
	flagSet.StringSliceP(azparams.ParamSpotExcludedRegions, "", []string{}, azparams.ParamSpotExcludedRegionsDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	params.AddCirrusFlags(flagSet)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
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
			if err := azureWindows.Destroy(
				&maptContext.ContextArgs{
					ProjectName: viper.GetString(params.ProjectName),
					BackedURL:   viper.GetString(params.BackedURL),
					Debug:       viper.IsSet(params.Debug),
					DebugLevel:  viper.GetUint(params.DebugLevel),
				}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
