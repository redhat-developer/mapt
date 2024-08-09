package hosts

import (
	"fmt"

	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"

	spotprice "github.com/redhat-developer/mapt/pkg/provider/azure/module/spot-price"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdUbuntu     = "ubuntu"
	cmdUbuntuDesc = "ubuntu operations"
	cmdRHEL       = "rhel"
	cmdRHELDesc   = "ubuntu operations"

	paramLinuxVersion     = "version"
	paramLinuxVersionDesc = "linux version. Version should be formmated as X.Y (Major.minor)"
	defaultUbuntuVersion  = "24.04"
	defaultRHELVersion    = "9.4"
)

func GetUbuntuCmd() *cobra.Command {
	return getLinuxCmd(cmdUbuntu, cmdUbuntuDesc, azureLinux.Ubuntu, defaultUbuntuVersion)
}

func GetRHELCmd() *cobra.Command {
	return getLinuxCmd(cmdRHEL, cmdRHELDesc, azureLinux.RHEL, defaultRHELVersion)
}

func getLinuxCmd(cmd, cmdDesc string, ostype azureLinux.OSType, defaultOSVersion string) *cobra.Command {
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
	c.AddCommand(getCreateLinux(ostype, defaultOSVersion), getDestroyLinux())
	return c
}

func getCreateLinux(ostype azureLinux.OSType, defaultOSVersion string) *cobra.Command {
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
			var spotToleranceValue = spotprice.DefaultEvictionRate
			if viper.IsSet(paramSpotTolerance) {
				var ok bool
				spotToleranceValue, ok = spotprice.ParseEvictionRate(
					viper.GetString(paramSpotTolerance))
				if !ok {
					return fmt.Errorf("%s is not a valid spot tolerance value", viper.GetString(paramSpotTolerance))
				}
			}
			if err := azureLinux.Create(
				&azureLinux.LinuxRequest{
					Prefix:        viper.GetString(params.ProjectName),
					Location:      viper.GetString(paramLocation),
					VMSize:        viper.GetString(paramVMSize),
					Version:       viper.GetString(paramLinuxVersion),
					Arch:          viper.GetString(params.LinuxArch),
					OSType:        ostype,
					Username:      viper.GetString(paramUsername),
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
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringP(paramVMSize, "", defaultVMSize, paramVMSizeDesc)
	flagSet.StringP(paramLinuxVersion, "", defaultOSVersion, paramLinuxVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.Bool(paramSpot, false, paramSpotDesc)
	flagSet.StringP(paramSpotTolerance, "", defaultSpotTolerance, paramSpotToleranceDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroyLinux() *cobra.Command {
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
			if err := azureLinux.Destroy(); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
