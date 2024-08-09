package hosts

import (
	"fmt"

	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/constants"
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
			if viper.IsSet(azparams.ParamSpotTolerance) {
				var ok bool
				spotToleranceValue, ok = spotprice.ParseEvictionRate(
					viper.GetString(azparams.ParamSpotTolerance))
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

			if err := azureUbuntu.Create(
				&azureUbuntu.UbuntuRequest{
					Prefix:        viper.GetString(params.ProjectName),
					Location:      viper.GetString(azparams.ParamLocation),
					VMSize:        viper.GetString(azparams.ParamVMSize),
					Version:       viper.GetString(paramUbuntuVersion),
					Username:      viper.GetString(paramUsername),
					Spot:          viper.IsSet(azparams.ParamSpot),
					SpotTolerance: spotToleranceValue}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(azparams.ParamLocation, "", azparams.DefaultLocation, azparams.ParamLocationDesc)
	flagSet.StringP(azparams.ParamVMSize, "", azparams.DefaultVMSize, azparams.ParamVMSizeDesc)
	flagSet.StringP(paramUbuntuVersion, "", defaultUbuntuVersion, paramUbuntuVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.Bool(azparams.ParamSpot, false, azparams.ParamSpotDesc)
	flagSet.StringP(azparams.ParamSpotTolerance, "", azparams.DefaultSpotTolerance, azparams.ParamSpotToleranceDesc)
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
