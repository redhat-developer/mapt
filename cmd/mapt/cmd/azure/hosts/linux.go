package hosts

import (
	azureParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/params"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureLinux "github.com/redhat-developer/mapt/pkg/provider/azure/action/linux"
	"github.com/redhat-developer/mapt/pkg/provider/azure/data"

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
			return azureLinux.Create(
				&maptContext.ContextArgs{
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&azureLinux.LinuxArgs{
					ComputeRequest: params.ComputeRequestArgs(),
					Spot:           params.SpotArgs(),
					Location:       viper.GetString(azureParams.Location),
					Version:        viper.GetString(paramLinuxVersion),
					Arch:           viper.GetString(params.LinuxArch),
					OSType:         ostype,
					Username:       viper.GetString(paramUsername)})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(azureParams.Location, "", azureParams.LocationDefault, azureParams.LocationDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringP(paramLinuxVersion, "", defaultOSVersion, paramLinuxVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
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
			return azureLinux.Destroy(
				&maptContext.ContextArgs{
					ProjectName: viper.GetString(params.ProjectName),
					BackedURL:   viper.GetString(params.BackedURL),
					Debug:       viper.IsSet(params.Debug),
					DebugLevel:  viper.GetUint(params.DebugLevel),
				})
		},
	}
}
