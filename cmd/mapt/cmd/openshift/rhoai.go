package openshift

import (
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	rhoaiAction "github.com/redhat-developer/mapt/pkg/provider/openshift/action/rhoai"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdRHOAI     = "rhoai"
	cmdRHOAIDesc = "deploy Red Hat OpenShift AI on an existing OpenShift cluster"

	kubeconfig     = "kubeconfig"
	kubeconfigDesc = "path to the kubeconfig file for the target OpenShift cluster"

	rhoaiProfile     = "profile"
	rhoaiProfileDesc = "profiles to deploy (ai, virtualization, serverless-serving, serverless-eventing, serverless, servicemesh, nvidia)"
)

func getRHOAICmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdRHOAI,
		Short: cmdRHOAIDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(createRHOAI(), destroyRHOAI())
	return c
}

func createRHOAI() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := viper.BindPFlags(cmd.InheritedFlags()); err != nil {
				return err
			}
			profiles := viper.GetStringSlice(rhoaiProfile)
			if len(profiles) == 0 {
				profiles = []string{"ai"}
			}
			return rhoaiAction.Create(
				&maptContext.ContextArgs{
					Context:     cmd.Context(),
					ProjectName: viper.GetString(params.ProjectName),
					BackedURL:   viper.GetString(params.BackedURL),
					Debug:       viper.IsSet(params.Debug),
					DebugLevel:  viper.GetUint(params.DebugLevel),
					Tags:        viper.GetStringMapString(params.Tags),
				},
				&rhoaiAction.RHOAIArgs{
					KubeconfigPath: viper.GetString(kubeconfig),
					Profiles:       profiles,
				})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(kubeconfig, "", "", kubeconfigDesc)
	flagSet.StringSliceP(rhoaiProfile, "", []string{}, rhoaiProfileDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	_ = c.MarkFlagRequired(kubeconfig)
	return c
}

func destroyRHOAI() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := viper.BindPFlags(cmd.InheritedFlags()); err != nil {
				return err
			}
			return rhoaiAction.Destroy(&maptContext.ContextArgs{
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
