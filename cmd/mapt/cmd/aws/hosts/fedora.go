package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/fedora"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdFedora     = "fedora"
	cmdFedoraDesc = "manage fedora dedicated host"

	fedoraVersion        string = "version"
	fedoraVersionDesc    string = "version for the Fedora Cloud OS"
	fedoraVersionDefault string = "41"
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

	flagSet := pflag.NewFlagSet(cmdFedora, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

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
			return fedora.Create(
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
				&fedora.FedoraArgs{
					Prefix:         "main",
					Version:        viper.GetString(fedoraVersion),
					Arch:           viper.GetString(params.LinuxArch),
					ComputeRequest: params.ComputeRequestArgs(),
					Spot:           params.SpotArgs(),
					Timeout:        viper.GetString(params.Timeout),
					Airgap:         viper.IsSet(airgap)})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(fedoraVersion, "", fedoraVersionDefault, fedoraVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
	params.AddGHActionsFlags(flagSet)
	params.AddCirrusFlags(flagSet)
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
			return fedora.Destroy(&maptContext.ContextArgs{
				ProjectName:  viper.GetString(params.ProjectName),
				BackedURL:    viper.GetString(params.BackedURL),
				Debug:        viper.IsSet(params.Debug),
				DebugLevel:   viper.GetUint(params.DebugLevel),
				Serverless:   viper.IsSet(params.Serverless),
				ForceDestroy: viper.IsSet(params.ForceDestroy),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.ForceDestroy, false, params.ForceDestroyDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
