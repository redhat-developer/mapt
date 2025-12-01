package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	rhelai "github.com/redhat-developer/mapt/pkg/provider/aws/action/rhel-ai"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdRHELAI     = "rhel-ai"
	cmdRHELAIDesc = "manage rhel ai host"
)

func GetRHELAICmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdRHELAI,
		Short: cmdRHELAIDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}

	flagSet := pflag.NewFlagSet(cmdRHELAI, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(getRHELAICreate(), getRHELAIDestroy())
	return c
}

func getRHELAICreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return rhelai.Create(
				&maptContext.ContextArgs{
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&rhelai.RHELAIArgs{
					Prefix:         "main",
					Version:        viper.GetString(params.RhelAIVersion),
					CustomAMI:      viper.GetString(params.RhelAIAMICustom),
					SubsUsername:   viper.GetString(params.SubsUsername),
					SubsUserpass:   viper.GetString(params.SubsUserpass),
					ComputeRequest: params.ComputeRequestArgs(),
					Spot:           params.SpotArgs(),
					Timeout:        viper.GetString(params.Timeout),
				})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(params.RhelAIVersion, "", params.RhelAIVersionDefault, params.RhelAIVersionDesc)
	flagSet.StringP(params.RhelAIAMICustom, "", "", params.RhelAIAMICustomDesc)
	flagSet.StringP(params.SubsUsername, "", "", params.SubsUsernameDesc)
	flagSet.StringP(params.SubsUserpass, "", "", params.SubsUserpassDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getRHELAIDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return rhelai.Destroy(&maptContext.ContextArgs{
				ProjectName:  viper.GetString(params.ProjectName),
				BackedURL:    viper.GetString(params.BackedURL),
				Debug:        viper.IsSet(params.Debug),
				DebugLevel:   viper.GetUint(params.DebugLevel),
				Serverless:   viper.IsSet(params.Serverless),
				ForceDestroy: viper.IsSet(params.ForceDestroy),
				CleanupState: viper.IsSet(params.CleanupState),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.ForceDestroy, false, params.ForceDestroyDesc)
	flagSet.Bool(params.CleanupState, true, params.CleanupStateDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
