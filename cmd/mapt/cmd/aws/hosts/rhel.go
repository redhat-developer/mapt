package hosts

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/rhel"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdRHEL     = "rhel"
	cmdRHELDesc = "manage rhel dedicated host"
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

	flagSet := pflag.NewFlagSet(cmdRHEL, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(getRHELCreate(), getRHELDestroy())
	return c
}

func getRHELCreate() *cobra.Command {
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
			if err := rhel.Create(
				ctx,
				&rhel.RHELArgs{
					Prefix:         "main",
					Version:        viper.GetString(params.RhelVersion),
					Arch:           viper.GetString(params.LinuxArch),
					ComputeRequest: params.GetComputeRequest(),
					SubsUsername:   viper.GetString(params.SubsUsername),
					SubsUserpass:   viper.GetString(params.SubsUserpass),
					ProfileSNC:     viper.IsSet(params.ProfileSNC),
					Spot:           viper.IsSet(awsParams.Spot),
					Timeout:        viper.GetString(params.Timeout),
					Airgap:         viper.IsSet(airgap),
				}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(params.RhelVersion, "", params.RhelVersionDefault, params.RhelVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringP(params.SubsUsername, "", "", params.SubsUsernameDesc)
	flagSet.StringP(params.SubsUserpass, "", "", params.SubsUserpassDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.Bool(awsParams.Spot, false, awsParams.SpotDesc)
	flagSet.IntP(params.SpotPriceIncreaseRate, "", params.SpotPriceIncreaseRateDefault, params.SpotPriceIncreaseRateDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	flagSet.Bool(params.ProfileSNC, false, params.ProfileSNCDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	params.AddCirrusFlags(flagSet)
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getRHELDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			if err := rhel.Destroy(&maptContext.ContextArgs{
				ProjectName: viper.GetString(params.ProjectName),
				BackedURL:   viper.GetString(params.BackedURL),
				Debug:       viper.IsSet(params.Debug),
				DebugLevel:  viper.GetUint(params.DebugLevel),
				Serverless:  viper.IsSet(params.Serverless),
			}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
