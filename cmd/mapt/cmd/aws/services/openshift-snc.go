package services

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	openshiftsnc "github.com/redhat-developer/mapt/pkg/provider/aws/action/openshift-snc"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdOpenshiftSNC     = "openshift-snc"
	cmdOpenshiftSNCDesc = "Manage an OpenShift Single Node Cluster based on OpenShift Local. This is not intended for production use"

	ocpVersion         = "version"
	ocpVersionDesc     = "version for Openshift. If not set it will pick latest available version"
	pullSecretFile     = "pull-secret-file"
	pullSecretFileDesc = "file path of image pull secret (download from https://console.redhat.com/openshift/create/local)"
	caCertFile         = "ca-cert-file"
	caCertFileDesc     = "file path for Certified Authority to sign for certs on cluster, as requisite the CA subj CN should contain /CN=admin-kubeconfig-signer-custom. If no ca.crt is given mapt will create one"
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
			if err := openshiftsnc.Create(
				&maptContext.ContextArgs{
					ProjectName:           viper.GetString(params.ProjectName),
					BackedURL:             viper.GetString(params.BackedURL),
					ResultsOutput:         viper.GetString(params.ConnectionDetailsOutput),
					Debug:                 viper.IsSet(params.Debug),
					DebugLevel:            viper.GetUint(params.DebugLevel),
					SpotPriceIncreaseRate: viper.GetInt(params.SpotPriceIncreaseRate),
					Tags:                  viper.GetStringMapString(params.Tags),
				},
				&openshiftsnc.OpenshiftSNCArgs{
					InstanceRequest: &instancetypes.AwsInstanceRequest{
						CPUs:      viper.GetInt32(params.CPUs),
						MemoryGib: viper.GetInt32(params.Memory),
						Arch: util.If(viper.GetString(params.LinuxArch) == "arm64",
							instancetypes.Arm64, instancetypes.Amd64),
						NestedVirt: viper.GetBool(params.ProfileSNC) || viper.GetBool(params.NestedVirt),
					},
					Version:        viper.GetString(ocpVersion),
					Arch:           viper.GetString(params.LinuxArch),
					PullSecretFile: viper.GetString(pullSecretFile),
					CaCertFile:     viper.GetString(caCertFile),
					Spot:           viper.IsSet(awsParams.Spot),
					Timeout:        viper.GetString(params.Timeout)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringP(ocpVersion, "", "", ocpVersionDesc)
	flagSet.StringP(params.LinuxArch, "", params.LinuxArchDefault, params.LinuxArchDesc)
	flagSet.StringP(pullSecretFile, "", "", pullSecretFileDesc)
	flagSet.StringP(caCertFile, "", "", caCertFileDesc)
	flagSet.Bool(awsParams.Spot, false, awsParams.SpotDesc)
	flagSet.IntP(params.SpotPriceIncreaseRate, "", params.SpotPriceIncreaseRateDefault, params.SpotPriceIncreaseRateDesc)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	flagSet.AddFlagSet(params.GetCpusAndMemoryFlagset())
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
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

			if err := openshiftsnc.Destroy(&maptContext.ContextArgs{
				ProjectName: viper.GetString(params.ProjectName),
				BackedURL:   viper.GetString(params.BackedURL),
				Debug:       viper.IsSet(params.Debug),
				DebugLevel:  viper.GetUint(params.DebugLevel),
				Serverless:  viper.IsSet(params.Serverless),
			}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.DestroyCmdName, pflag.ExitOnError)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
