package hosts

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/gcp/action/windows"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdWindows     = "windows"
	cmdWindowsDesc = "manage windows on gcp"

	windowsVersion        = "windows-version"
	windowsVersionDesc    = "Windows Server version (e.g. 2022, 2019)"
	windowsVersionDefault = "2022"
)

func GetWindowsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdWindows,
		Short: cmdWindowsDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}

	flagSet := pflag.NewFlagSet(cmdWindows, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(getWindowsCreate(), getWindowsDestroy())
	return c
}

func getWindowsCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return windows.Create(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&windows.WindowsArgs{
					Prefix:         "main",
					Version:        viper.GetString(windowsVersion),
					ComputeRequest: params.ComputeRequestArgs(),
					Spot:           params.SpotArgs(),
				})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(windowsVersion, "", windowsVersionDefault, windowsVersionDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getWindowsDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return windows.Destroy(&maptContext.ContextArgs{
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
