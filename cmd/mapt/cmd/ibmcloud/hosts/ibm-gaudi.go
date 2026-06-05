package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	ibmgaudi "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/action/ibm-gaudi"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdIBMGaudi     = "ibm-gaudi"
	cmdIBMGaudiDesc = "manage ibm gaudi3 accelerated instances (amd64)"
)

func IBMGaudiCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdIBMGaudi,
		Short: cmdIBMGaudiDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}

	flagSet := pflag.NewFlagSet(cmdIBMGaudi, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(ibmGaudiCreate(), ibmGaudiDestroy())
	return c
}

func ibmGaudiCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return ibmgaudi.New(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&ibmgaudi.GaudiArgs{
					SubnetID:       viper.GetString(params.SubnetID),
					OtelAppCode:    viper.GetString(params.OtelAppCode),
					OtelAuthToken:  viper.GetString(params.OtelAuthToken),
					OtelEndpoint:   viper.GetString(params.OtelEndpoint),
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
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func ibmGaudiDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return ibmgaudi.Destroy(&maptContext.ContextArgs{
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
