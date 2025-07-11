package hosts

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/fedora"
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

			ctx := &maptContext.ContextArgs{
				ProjectName:           viper.GetString(params.ProjectName),
				BackedURL:             viper.GetString(params.BackedURL),
				ResultsOutput:         viper.GetString(params.ConnectionDetailsOutput),
				Debug:                 viper.IsSet(params.Debug),
				DebugLevel:            viper.GetUint(params.DebugLevel),
				SpotPriceIncreaseRate: viper.GetInt(params.SpotPriceIncreaseRate),
				Tags:                  viper.GetStringMapString(params.Tags),
			}

			if viper.IsSet(params.CirrusPWToken) {
				ctx.CirrusPWArgs = &cirrus.PersistentWorkerArgs{
					Token:    viper.GetString(params.CirrusPWToken),
					Labels:   viper.GetStringMapString(params.CirrusPWLabels),
					Platform: &cirrus.Linux,
					Arch: params.LinuxArchAsCirrusArch(
						viper.GetString(params.LinuxArch)),
				}
			}

			if viper.IsSet(params.GHActionsRunnerToken) {
				ctx.GHRunnerArgs = &github.GithubRunnerArgs{
					Token:    viper.GetString(params.GHActionsRunnerToken),
					RepoURL:  viper.GetString(params.GHActionsRunnerRepo),
					Labels:   viper.GetStringSlice(params.GHActionsRunnerLabels),
					Platform: &github.Linux,
					Arch: params.LinuxArchAsGithubActionsArch(
						viper.GetString(params.LinuxArch)),
				}
			}

			// Run create
			if err := fedora.Create(
				ctx,
				&fedora.FedoraArgs{
					Prefix:         "main",
					Version:        viper.GetString(fedoraVersion),
					Arch:           viper.GetString(params.LinuxArch),
					ComputeRequest: params.GetComputeRequest(),
					Spot:           viper.IsSet(awsParams.Spot),
					Timeout:        viper.GetString(params.Timeout),
					Airgap:         viper.IsSet(airgap)}); err != nil {
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
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.Bool(awsParams.Spot, false, awsParams.SpotDesc)
	flagSet.IntP(params.SpotPriceIncreaseRate, "", params.SpotPriceIncreaseRateDefault, params.SpotPriceIncreaseRateDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	params.AddCirrusFlags(flagSet)
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

			if err := fedora.Destroy(&maptContext.ContextArgs{
				ProjectName:  viper.GetString(params.ProjectName),
				BackedURL:    viper.GetString(params.BackedURL),
				Debug:        viper.IsSet(params.Debug),
				DebugLevel:   viper.GetUint(params.DebugLevel),
				Serverless:   viper.IsSet(params.Serverless),
				ForceDestroy: viper.IsSet(params.ForceDestroy),
			}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.ForceDestroy, false, params.ForceDestroyDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
