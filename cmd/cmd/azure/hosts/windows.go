package hosts

import (
	"fmt"

	params "github.com/adrianriobo/qenvs/cmd/cmd/constants"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	azureWindows "github.com/adrianriobo/qenvs/pkg/provider/azure/action/windows"
	spotprice "github.com/adrianriobo/qenvs/pkg/provider/azure/module/spot-price"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmd     = "windows"
	cmdDesc = "windows operations"

	paramLocation          = "location"
	paramLocationDesc      = "location for created resources within Windows desktop"
	defaultLocation        = "West US"
	paramVMSize            = "vmsize"
	paramVMSizeDesc        = "size for the VM. Type requires to allow nested virtualization"
	defaultVMSize          = "Standard_D8s_v5"
	paramVersion           = "windows-version"
	paramVersionDesc       = "Major version for windows desktop 10 or 11"
	defaultVersion         = "11"
	paramFeature           = "windows-featurepack"
	paramFeatureDesc       = "windows feature pack"
	defaultFeature         = "23h2-pro"
	paramUsername          = "username"
	paramUsernameDesc      = "username for general user. SSH accessible + rdp with generated password"
	defaultUsername        = "rhqp"
	paramAdminUsername     = "admin-username"
	paramAdminUsernameDesc = "username for admin user. Only rdp accessible within generated password"
	defaultAdminUsername   = "rhqpadmin"
	paramSpot              = "spot"
	paramSpotDesc          = "if spot is set the spot prices across all regions will be cheked and machine will be started on best spot option (price / eviction)"
	paramSpotTolerance     = "spot-eviction-tolerance"
	paramSpotToleranceDesc = "if spot is enable we can define the minimum tolerance level of eviction. Allowed value are: lowest, low, medium, high or highest"
	defaultSpotTolerance   = "lowest"
	paramProfile           = "profile"
	paramProfileDesc       = "comma seperated list of profiles to apply on the target machine. Profiles available: crc"
)

func GetWindowsDesktopCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmd,
		Short: cmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCreate(), getDestroy())
	return c
}

func getCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize context
			qenvsContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags))

			var spotToleranceValue = spotprice.DefaultEvictionRate
			if viper.IsSet(paramSpotTolerance) {
				var ok bool
				spotToleranceValue, ok = spotprice.ParseEvictionRate(
					viper.GetString(paramSpotTolerance))
				if !ok {
					return fmt.Errorf("%s is not a valid spot tolerance value", viper.GetString(paramSpotTolerance))
				}
			}
			// ParseEvictionRate
			if err := azureWindows.Create(
				&azureWindows.WindowsRequest{
					Prefix:        "",
					Location:      viper.GetString(paramLocation),
					VMSize:        viper.GetString(paramVMSize),
					Version:       viper.GetString(paramVersion),
					Feature:       viper.GetString(paramFeature),
					Username:      viper.GetString(paramUsername),
					AdminUsername: viper.GetString(paramAdminUsername),
					Profiles:      viper.GetStringSlice(paramProfile),
					Spot:          viper.IsSet(paramSpot),
					SpotTolerance: spotToleranceValue}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramLocation, "", defaultLocation, paramLocationDesc)
	flagSet.StringP(paramVMSize, "", defaultVMSize, paramVMSizeDesc)
	flagSet.StringP(paramVersion, "", defaultVersion, paramVersionDesc)
	flagSet.StringP(paramFeature, "", defaultFeature, paramFeatureDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.StringP(paramAdminUsername, "", defaultAdminUsername, paramAdminUsernameDesc)
	flagSet.StringSliceP(paramProfile, "", []string{}, paramProfileDesc)
	flagSet.Bool(paramSpot, false, paramSpotDesc)
	flagSet.StringP(paramSpotTolerance, "", defaultSpotTolerance, paramSpotToleranceDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroy() *cobra.Command {
	return &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize context
			qenvsContext.Init(
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
