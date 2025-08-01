package hosts

import (
	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/params"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureRHEL "github.com/redhat-developer/mapt/pkg/provider/azure/action/rhel"

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

			if err := azureRHEL.Create(
				ctx,
				&azureRHEL.RhelArgs{
					Location:            viper.GetString(paramLocation),
					ComputeRequest:      params.GetComputeRequest(),
					Version:             viper.GetString(paramLinuxVersion),
					Arch:                viper.GetString(params.LinuxArch),
					SubsUsername:        viper.GetString(params.SubsUsername),
					SubsUserpass:        viper.GetString(params.SubsUserpass),
					ProfileSNC:          viper.IsSet(params.ProfileSNC),
					Username:            viper.GetString(paramUsername),
					Spot:                viper.IsSet(azparams.ParamSpot),
					SpotTolerance:       *spotToleranceValue,
					SpotExcludedRegions: viper.GetStringSlice(azparams.ParamSpotExcludedRegions),
				}); err != nil {
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
	flagSet.StringP(paramLinuxVersion, "", defaultRHELVersion, paramLinuxVersionDesc)
	flagSet.StringP(paramUsername, "", defaultUsername, paramUsernameDesc)
	flagSet.StringP(params.SubsUsername, "", "", params.SubsUsernameDesc)
	flagSet.StringP(params.SubsUserpass, "", "", params.SubsUserpassDesc)
	flagSet.Bool(params.ProfileSNC, false, params.ProfileSNCDesc)
	flagSet.Bool(azparams.ParamSpot, false, azparams.ParamSpotDesc)
	flagSet.StringP(azparams.ParamSpotTolerance, "", azparams.DefaultSpotTolerance, azparams.ParamSpotToleranceDesc)
	flagSet.StringSliceP(azparams.ParamSpotExcludedRegions, "", []string{}, azparams.ParamSpotExcludedRegionsDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	params.AddCirrusFlags(flagSet)
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
			if err := azureRHEL.Destroy(
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
