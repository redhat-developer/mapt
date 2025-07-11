package hosts

import (
	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/params"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"

	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdUbuntu     = "ubuntu"
	cmdUbuntuDesc = "ubuntu operations"
	cmdFedora     = "fedora"
	cmdFedoraDesc = "fedora operations"
)

func GetUbuntuCmd() *cobra.Command {
	return getLinuxCmd(cmdUbuntu, cmdUbuntuDesc, data.Ubuntu, defaultUbuntuVersion)
}

func GetFedoraCmd() *cobra.Command {
	return getLinuxCmd(cmdFedora, cmdFedoraDesc, data.Fedora, defaultFedoraVersion)
}

func getLinuxCmd(cmd, cmdDesc string, ostype data.OSType, defaultOSVersion string) *cobra.Command {
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

func getCreateLinux(ostype data.OSType, defaultOSVersion string) *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			spotToleranceValue, err := azparams.SpotTolerance()
			if err != nil {
				return err
			}

			ctx := &maptContext.ContextArgs{
				ProjectName:   viper.GetString(params.ProjectName),
				BackedURL:     viper.GetString(params.BackedURL),
				ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
				Debug:         viper.IsSet(params.Debug),
				DebugLevel:    viper.GetUint(params.DebugLevel),
				Tags:          viper.GetStringMapString(params.Tags),
			}

			if err := azureLinux.Create(
				ctx,
				&azureLinux.LinuxRequest{
					Prefix:         viper.GetString(params.ProjectName),
					Location:       viper.GetString(paramLocation),
					VMSizes:        viper.GetStringSlice(paramVMSize),
					ComputeRequest: params.GetComputeRequest(),
					Version:        viper.GetString(paramLinuxVersion),
					Arch:           viper.GetString(params.LinuxArch),
					OSType:         ostype,
					Username:       viper.GetString(paramUsername),
					Spot:           viper.IsSet(azparams.ParamSpot),
					SpotTolerance:  *spotToleranceValue}); err != nil {
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
	flagSet.StringSliceP(paramVMSize, "", []string{}, paramVMSizeDesc)
	flagSet.StringP(paramLinuxVersion, "", defaultOSVersion, paramLinuxVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.Bool(azparams.ParamSpot, false, azparams.ParamSpotDesc)
	flagSet.StringP(azparams.ParamSpotTolerance, "", azparams.DefaultSpotTolerance, azparams.ParamSpotToleranceDesc)
	flagSet.StringSliceP(azparams.ParamSpotExcludedRegions, "", []string{}, azparams.ParamSpotExcludedRegionsDesc)
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
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
			if err := azureLinux.Destroy(
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
