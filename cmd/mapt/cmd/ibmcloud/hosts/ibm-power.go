package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmpower "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/action/ibm-power"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdIBMPower     = "ibm-power"
	cmdIBMPowerDesc = "manage ibm-power machines (ppc64)"
)

func IBMPowerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdIBMPower,
		Short: cmdIBMPowerDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}

	flagSet := pflag.NewFlagSet(cmdIBMPower, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(ibmPowerCreate(), ibmPowerDestroy())
	return c
}

func ibmPowerCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return ibmpower.New(
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
				&ibmpower.PWArgs{
					PIPrivateSubnetID: viper.GetString(params.PIPrivateSubnetID),
					WorkspaceID:       viper.GetString(params.WorkspaceID),
					VPCPublicSubnetID: viper.GetString(params.VPCPublicSubnetID),
					OtelAppCode:       viper.GetString(params.OtelAppCode),
					OtelAuthToken:     viper.GetString(params.OtelAuthToken),
					OtelEndpoint:      viper.GetString(params.OtelEndpoint),
					OtelIndex:          viper.GetString(params.OtelIndex),
					OtelExtraAttrs:     viper.GetStringMapString(params.OtelExtraAttrs),
				})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(params.PIPrivateSubnetID, "", "", params.PIPrivateSubnetIDDesc)
	flagSet.StringP(params.WorkspaceID, "", "", params.WorkspaceIDDesc)
	flagSet.StringP(params.VPCPublicSubnetID, "", "", params.VPCPublicSubnetIDDesc)
	flagSet.StringP(params.OtelAppCode, "", "", params.OtelAppCodeDesc)
	flagSet.StringP(params.OtelAuthToken, "", "", params.OtelAuthTokenDesc)
	flagSet.StringP(params.OtelEndpoint, "", "https://otel-input.corp.redhat.com", params.OtelEndpointDesc)
	flagSet.StringP(params.OtelIndex, "", "", params.OtelIndexDesc)
	flagSet.StringToStringP(params.OtelExtraAttrs, "", nil, params.OtelExtraAttrsDesc)
	params.AddGHActionsFlags(flagSet)
	params.AddCirrusFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	_ = c.MarkPersistentFlagRequired(params.PIPrivateSubnetID)
	_ = c.MarkPersistentFlagRequired(params.WorkspaceID)
	return c
}

func ibmPowerDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return ibmpower.Destroy(&maptContext.ContextArgs{
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
