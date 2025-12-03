package services

import (
	"fmt"

	azureParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/params"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/azure/action/kind"
	kindApi "github.com/redhat-developer/mapt/pkg/targets/service/kind"
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
	flagSet := pflag.NewFlagSet(params.KindCmd, pflag.ExitOnError)
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

			// Parse extra port mappings from JSON string to PortMapping struct
			var extraPortMappings []kindApi.PortMapping
			extraPortMappingsStr := viper.GetString(params.KindExtraPortMappings)
			if extraPortMappingsStr != "" {
				var err error
				extraPortMappings, err = kindApi.ParseExtraPortMappings(extraPortMappingsStr)
				if err != nil {
					return fmt.Errorf("failed to parse 'extra-port-mappings' flag: %w", err)
				}
			}

			if _, err := kind.Create(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&kindApi.KindArgs{
					ComputeRequest:    params.ComputeRequestArgs(),
					Spot:              params.SpotArgs(),
					HostingPlace:      viper.GetString(azureParams.Location),
					Version:           viper.GetString(params.KindK8SVersion),
					Arch:              viper.GetString(params.LinuxArch),
					ExtraPortMappings: extraPortMappings}); err != nil {
				return err
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringP(params.KindK8SVersion, "", params.KindK8SVersionDefault, params.KindK8SVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringP(params.KindExtraPortMappings, "", "", params.KindExtraPortMappingsDesc)
	flagSet.StringP(azureParams.Location, "", azureParams.LocationDefault, azureParams.LocationDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
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
			return kind.Destroy(&maptContext.ContextArgs{
				Context:       cmd.Context(),
				ProjectName:   viper.GetString(params.ProjectName),
				BackedURL:     viper.GetString(params.BackedURL),
				Debug:         viper.IsSet(params.Debug),
				DebugLevel:    viper.GetUint(params.DebugLevel),
				Serverless:    viper.IsSet(params.Serverless),
				ForceDestroy:  viper.IsSet(params.ForceDestroy),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.ForceDestroy, false, params.ForceDestroyDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
