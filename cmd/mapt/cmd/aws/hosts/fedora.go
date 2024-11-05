package hosts

import (
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/fedora"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdFedora     = "fedora"
	cmdFedoraDesc = "manage fedora dedicated host"

	fedoraVersion        string = "version"
	fedoraVersionDesc    string = "version for the Fedora Cloud OS"
	fedoraVersionDefault string = "40"
)

func GetFedoraCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdFedora,
		Short: cmdFedoraDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getFedoraCreate(), getFedoraDestroy())
	return c
}

func getFedoraCreate() *cobra.Command {
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

			// Initialize gh actions runner if needed
			if viper.IsSet(params.InstallGHActionsRunner) {
				err := ghactions.InitGHRunnerArgs(viper.GetString(params.GHActionsRunnerToken),
					viper.GetString(params.GHActionsRunnerName),
					viper.GetString(params.GHActionsRunnerRepo),
					viper.GetStringSlice(params.GHActionsRunnerLabels))
				if err != nil {
					logging.Fatal(err)
				}
			}
			instanceRequest := &instancetypes.AwsInstanceRequest{
				CPUs:       viper.GetInt32(params.CPUs),
				MemoryGib:  viper.GetInt32(params.Memory),
				Arch:       util.If(viper.GetString(params.LinuxArch) == "arm64", instancetypes.Arm64, instancetypes.Amd64),
				NestedVirt: viper.GetBool(profileSNC) || viper.GetBool(params.NestedVirt),
			}

			// Run create
			if err := fedora.Create(
				&fedora.Request{
					Prefix:               "main",
					Version:              viper.GetString(rhelVersion),
					Arch:                 viper.GetString(params.LinuxArch),
					VMType:               viper.GetStringSlice(vmTypes),
					InstanceRequest:      instanceRequest,
					Spot:                 viper.IsSet(spot),
					SetupGHActionsRunner: viper.IsSet(params.InstallGHActionsRunner),
					Airgap:               viper.IsSet(airgap)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(fedoraVersion, "", fedoraVersionDefault, fedoraVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringSliceP(vmTypes, "", []string{}, vmTypesDescription)
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.Bool(spot, false, spotDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getFedoraDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			maptContext.InitBase(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL))

			if err := fedora.Destroy(); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	return c
}
