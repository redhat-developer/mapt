package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/gcp/action/rhel"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdRHEL     = "rhel"
	cmdRHELDesc = "manage rhel on gcp"
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
			return rhel.Create(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&rhel.RHELArgs{
					Prefix:         "main",
					Version:        viper.GetString(params.RhelVersion),
					Arch:           viper.GetString(params.LinuxArch),
					ComputeRequest: params.ComputeRequestArgs(),
					Spot:           params.SpotArgs(),
				})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(params.RhelVersion, "", params.RhelVersionDefault, params.RhelVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
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
			return rhel.Destroy(&maptContext.ContextArgs{
				Context:      cmd.Context(),
				ProjectName:  viper.GetString(params.ProjectName),
				BackedURL:    viper.GetString(params.BackedURL),
				Debug:        viper.IsSet(params.Debug),
				DebugLevel:   viper.GetUint(params.DebugLevel),
				ForceDestroy: viper.IsSet(params.ForceDestroy),
				KeepState:    viper.IsSet(params.KeepState),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.ForceDestroy, false, params.ForceDestroyDesc)
	flagSet.Bool(params.KeepState, false, params.KeepStateDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
