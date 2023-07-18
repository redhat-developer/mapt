package host

import (
	params "github.com/adrianriobo/qenvs/cmd/cmd/constants"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/environment"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmd     = "host"
	cmdDesc = "replica operations"
)

func GetCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmd,
		Short: cmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCreate(), getDestroy())
	return c
}

func getCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := environment.Create(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				// fixed as public to ensure sync, when PR is merged we can offer as param
				// https://github.com/pulumi/pulumi-command/pull/132
				true,
				viper.GetString(params.SupportedHostID),
				viper.GetString(params.RHMajorVersion),
				viper.GetString(params.RHSubcriptionUsername),
				viper.GetString(params.RHSubcriptionPassword),
				viper.GetString(params.FedoraMajorVersion),
				viper.GetString(params.MacOSMajorVersion)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringP(params.SupportedHostID, "", "", params.SupportedHostIDDesc)
	flagSet.StringP(params.RHMajorVersion, "", "8", params.RHMajorVersionDesc)
	flagSet.StringP(params.FedoraMajorVersion, "", "38", params.FedoraMajorVersionDesc)
	flagSet.StringP(params.RHSubcriptionUsername, "", "", params.RHSubcriptionUsernameDesc)
	flagSet.StringP(params.RHSubcriptionPassword, "", "", params.RHSubcriptionPasswordDesc)
	flagSet.StringP(params.MacOSMajorVersion, "", "13", params.MacOSMajorVersionDesc)

	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroy() *cobra.Command {
	return &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := environment.Destroy(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
