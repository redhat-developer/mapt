package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureWindows "github.com/redhat-developer/mapt/pkg/provider/azure/action/windows"
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
			return azureWindows.Create(
				&maptContext.ContextArgs{
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					CirrusPWArgs:  params.CirrusPersistentWorkerArgs(),
					GHRunnerArgs:  params.GithubRunnerArgs(),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&azureWindows.WindowsArgs{
					ComputeRequest: params.ComputeRequestArgs(),
					Spot:           params.SpotArgs(),
					Prefix:         viper.GetString(params.ProjectName),
					Location:       viper.GetString(paramLocation),
					Version:        viper.GetString(paramWindowsVersion),
					Feature:        viper.GetString(paramFeature),
					Username:       viper.GetString(paramUsername),
					AdminUsername:  viper.GetString(paramAdminUsername),
					Profiles:       viper.GetStringSlice(paramProfile)})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramLocation, "", defaultLocation, paramLocationDesc)
	flagSet.StringP(paramWindowsVersion, "", defaultWindowsVersion, paramWindowsVersionDesc)
	flagSet.StringP(paramFeature, "", defaultFeature, paramFeatureDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.StringP(paramAdminUsername, "", defaultAdminUsername, paramAdminUsernameDesc)
	flagSet.StringSliceP(paramProfile, "", []string{}, paramProfileDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
	params.AddGHActionsFlags(flagSet)
	params.AddCirrusFlags(flagSet)
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
