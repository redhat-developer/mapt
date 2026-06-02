package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmz "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/action/ibm-z"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdIBMZ     = "ibm-z"
	cmdIBMZDesc = "manage ibm-z machines (s390x)"
)

func IBMZCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdIBMZ,
		Short: cmdIBMZDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}

	flagSet := pflag.NewFlagSet(cmdIBMZ, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(ibmZCreate(), ibmZDestroy())
	return c
}

func ibmZCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return ibmz.New(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					CirrusPWArgs:  params.CirrusPersistentWorkerArgs(),
					GHRunnerArgs:  params.GithubRunnerArgs(),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&ibmz.ZArgs{
					SubnetID:      viper.GetString(params.SubnetID),
					OtelAppCode:   viper.GetString(params.OtelAppCode),
					OtelAuthToken: viper.GetString(params.OtelAuthToken),
					OtelEndpoint:  viper.GetString(params.OtelEndpoint),
					OtelIndex:      viper.GetString(params.OtelIndex),
					OtelExtraAttrs: viper.GetStringMapString(params.OtelExtraAttrs),
				})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(params.SubnetID, "", "", params.SubnetIDDesc)
	flagSet.StringP(params.OtelAppCode, "", "", params.OtelAppCodeDesc)
	flagSet.StringP(params.OtelAuthToken, "", "", params.OtelAuthTokenDesc)
	flagSet.StringP(params.OtelEndpoint, "", "https://otel-input.corp.redhat.com", params.OtelEndpointDesc)
	flagSet.StringP(params.OtelIndex, "", "", params.OtelIndexDesc)
	flagSet.StringToStringP(params.OtelExtraAttrs, "", nil, params.OtelExtraAttrsDesc)
	params.AddGHActionsFlags(flagSet)
	params.AddCirrusFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func ibmZDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return ibmz.Destroy(&maptContext.ContextArgs{
				Context:      cmd.Context(),
				ProjectName:  viper.GetString(params.ProjectName),
				BackedURL:    viper.GetString(params.BackedURL),
				Debug:        viper.IsSet(params.Debug),
				DebugLevel:   viper.GetUint(params.DebugLevel),
				Serverless:   viper.IsSet(params.Serverless),
				ForceDestroy: viper.IsSet(params.ForceDestroy),
				KeepState:    viper.IsSet(params.KeepState),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.ForceDestroy, false, params.ForceDestroyDesc)
	flagSet.Bool(params.KeepState, false, params.KeepStateDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
