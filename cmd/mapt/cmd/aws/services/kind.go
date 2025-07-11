package services

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws/action/kind"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func GetKindCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   params.KindCmd,
		Short: params.KindCmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(cmdOpenshiftSNC, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	c.AddCommand(createKind(), destroyKind())
	return c

}

func createKind() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if _, err := kind.Create(
				&maptContext.ContextArgs{
					ProjectName:           viper.GetString(params.ProjectName),
					BackedURL:             viper.GetString(params.BackedURL),
					ResultsOutput:         viper.GetString(params.ConnectionDetailsOutput),
					Debug:                 viper.IsSet(params.Debug),
					DebugLevel:            viper.GetUint(params.DebugLevel),
					SpotPriceIncreaseRate: viper.GetInt(params.SpotPriceIncreaseRate),
					Tags:                  viper.GetStringMapString(params.Tags),
				},
				&kind.KindArgs{
					ComputeRequest: params.GetComputeRequest(),
					Version:        viper.GetString(params.KindK8SVersion),
					Arch:           viper.GetString(params.LinuxArch),
					Spot:           viper.IsSet(awsParams.Spot),
					Timeout:        viper.GetString(params.Timeout)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringP(params.KindK8SVersion, "", "", params.KindK8SVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.Bool(awsParams.Spot, false, awsParams.SpotDesc)
	flagSet.IntP(params.SpotPriceIncreaseRate, "", params.SpotPriceIncreaseRateDefault, params.SpotPriceIncreaseRateDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func destroyKind() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			if err := kind.Destroy(&maptContext.ContextArgs{
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
