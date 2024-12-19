package hosts

import (
	"fmt"

	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureRHEL "github.com/redhat-developer/mapt/pkg/provider/azure/action/rhel"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/util"

	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdRHEL     = "rhel"
	cmdRHELDesc = "RHEL operations"
)

func GetRHELCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdRHEL,
		Short: cmdRHELDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCreateRHEL(), getDestroyRHEL())
	return c
}

func getCreateRHEL() *cobra.Command {
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
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel))

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
			instanceRequest := &instancetypes.AzureInstanceRequest{
				CPUs:       viper.GetInt32(params.CPUs),
				MemoryGib:  viper.GetInt32(params.Memory),
				Arch:       util.If(viper.GetString(params.LinuxArch) == "arm64", instancetypes.Arm64, instancetypes.Amd64),
				NestedVirt: viper.GetBool(params.ProfileSNC) || viper.GetBool(params.NestedVirt),
			}

			// Initialize gh actions runner if needed
			if viper.IsSet(params.InstallGHActionsRunner) {
				err := ghactions.InitGHRunnerArgs(viper.GetString(params.GHActionsRunnerToken),
					viper.GetString(params.GHActionsRunnerName),
					viper.GetString(params.GHActionsRunnerRepo),
					viper.GetStringSlice(params.GHActionsRunnerLabels))
				if err != nil {
					logging.Error(err)
				}
			}

			if err := azureRHEL.Create(
				&azureRHEL.Request{
					Prefix:               viper.GetString(params.ProjectName),
					Location:             viper.GetString(paramLocation),
					VMSizes:              viper.GetStringSlice(paramVMSize),
					InstanceRequest:      instanceRequest,
					Version:              viper.GetString(paramLinuxVersion),
					Arch:                 viper.GetString(params.LinuxArch),
					SubsUsername:         viper.GetString(params.SubsUsername),
					SubsUserpass:         viper.GetString(params.SubsUserpass),
					ProfileSNC:           viper.IsSet(params.ProfileSNC),
					Username:             viper.GetString(paramUsername),
					Spot:                 viper.IsSet(paramSpot),
					SpotTolerance:        spotToleranceValue,
					SetupGHActionsRunner: viper.IsSet(params.InstallGHActionsRunner)}); err != nil {
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
	flagSet.StringP(paramLinuxVersion, "", defaultRHELVersion, paramLinuxVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.StringP(params.SubsUsername, "", "", params.SubsUsernameDesc)
	flagSet.StringP(params.SubsUserpass, "", "", params.SubsUserpassDesc)
	flagSet.Bool(params.ProfileSNC, false, params.ProfileSNCDesc)
	flagSet.Bool(paramSpot, false, paramSpotDesc)
	flagSet.StringP(paramSpotTolerance, "", defaultSpotTolerance, paramSpotToleranceDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroyRHEL() *cobra.Command {
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
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel))
			if err := azureRHEL.Destroy(); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
