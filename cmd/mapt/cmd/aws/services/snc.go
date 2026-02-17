package services

import (
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	openshiftsnc "github.com/redhat-developer/mapt/pkg/provider/aws/action/snc"
	sncApi "github.com/redhat-developer/mapt/pkg/target/service/snc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdOpenshiftSNC     = "openshift-snc"
	cmdOpenshiftSNCDesc = "Manage an OpenShift Single Node Cluster based on OpenShift Local. This is not intended for production use"

	ocpVersion        = "version"
	ocpDefaultVersion = "4.21.0"
	ocpVersionDesc    = "version for Openshift."

	pullSecretFile              = "pull-secret-file"
	pullSecretFileDesc          = "file path of image pull secret (download from https://console.redhat.com/openshift/create/local)"
	disableClusterReadiness     = "disable-cluster-readiness"
	disableClusterReadinessDesc = "If this flag is set it will skip the checks for the cluster readiness. In this case the kubeconfig can not be generated"
)

func GetOpenshiftSNCCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdOpenshiftSNC,
		Short: cmdOpenshiftSNCDesc,
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
	c.AddCommand(createSNC(), destroySNC())
	return c

}

func createSNC() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if _, err := openshiftsnc.Create(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&sncApi.SNCArgs{
					ComputeRequest:          params.ComputeRequestArgs(),
					Spot:                    params.SpotArgs(),
					Version:                 viper.GetString(ocpVersion),
					DisableClusterReadiness: viper.IsSet(disableClusterReadiness),
					Arch:                    viper.GetString(params.LinuxArch),
					PullSecretFile:          viper.GetString(pullSecretFile),
					Timeout:                 viper.GetString(params.Timeout)}); err != nil {
				return err
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringP(ocpVersion, "", ocpDefaultVersion, ocpVersionDesc)
	flagSet.Bool(disableClusterReadiness, false, disableClusterReadinessDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringP(pullSecretFile, "", "", pullSecretFileDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	params.AddComputeRequestFlags(flagSet)
	params.AddSpotFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func destroySNC() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return openshiftsnc.Destroy(&maptContext.ContextArgs{
				Context:     cmd.Context(),
				ProjectName: viper.GetString(params.ProjectName),
				BackedURL:   viper.GetString(params.BackedURL),
				Debug:       viper.IsSet(params.Debug),
				DebugLevel:  viper.GetUint(params.DebugLevel),
				Serverless:  viper.IsSet(params.Serverless),
				KeepState:   viper.IsSet(params.KeepState),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.KeepState, false, params.KeepStateDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
