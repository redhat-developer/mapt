package hosts

import (
	"fmt"

	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureUbuntu "github.com/redhat-developer/mapt/pkg/provider/azure/action/ubuntu"

	spotprice "github.com/redhat-developer/mapt/pkg/provider/azure/module/spot-price"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdUbunutu    = "ubuntu"
	cmdUbuntuDesc = "ubuntu operations"

	paramUbuntuVersion     = "version"
	paramUbuntuVersionDesc = "ubunutu version. Tore info at https://documentation.ubuntu.com/azure/en/latest/azure-how-to/instances/find-ubuntu-images"
	defaultUbuntuVersion   = "24_04"
)

func GetUbuntuCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdUbunutu,
		Short: cmdUbuntuDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCreateUbuntu(), getDestroyUbuntu())
	return c
}

func getCreateUbuntu() *cobra.Command {
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

			// Initialize gh actions runner if needed
			if viper.IsSet(params.InstallGHActionsRunner) {
				err := ghactions.InitGHRunnerArgs(viper.GetString(params.GHActionsRunnerToken),
					viper.GetString(params.GHActionsRunnerName),
					viper.GetString(params.GHActionsRunnerRepo))
				if err != nil {
					logging.Error(err)
				}
			}

			if err := azureUbuntu.Create(
				&azureUbuntu.UbuntuRequest{
					Prefix:        viper.GetString(params.ProjectName),
					Location:      viper.GetString(paramLocation),
					VMSize:        viper.GetString(paramVMSize),
					Version:       viper.GetString(paramUbuntuVersion),
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
	flagSet.StringP(paramVMSize, "", defaultVMSize, paramVMSizeDesc)
	flagSet.StringP(paramUbuntuVersion, "", defaultUbuntuVersion, paramUbuntuVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.Bool(paramSpot, false, paramSpotDesc)
	flagSet.StringP(paramSpotTolerance, "", defaultSpotTolerance, paramSpotToleranceDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroyUbuntu() *cobra.Command {
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
			if err := azureUbuntu.Destroy(); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
