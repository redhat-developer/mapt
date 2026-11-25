package services

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureAKS "github.com/redhat-developer/mapt/pkg/provider/azure/action/aks"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdAKS     = "aks"
	cmdAKSDesc = "aks operations"

	paramVersion              = "version"
	paramVersionDesc          = "AKS K8s cluster version"
	defaultVersion            = "1.31"
	paramOnlySystemPool       = "only-system-pool"
	paramOnlySystemPoolDesc   = "if we do not need bunch of resources we can run only the systempool. More info https://learn.microsoft.com/es-es/azure/aks/use-system-pools?tabs=azure-cli#system-and-user-node-pools"
	paramEnableAppRouting     = "enable-app-routing"
	paramEnableAppRoutingDesc = "enable application routing add-on with NGINX"

	paramLocation        = "location"
	paramLocationDesc    = "location for created resources in case spot flag (if available) is not passed"
	paramLocationDefault = "West US"
	paramVMSize          = "vmsize"
	paramVMSizeDesc      = "comma separated list of VMSizes to be used on the user pool. When spot is set the best option across the sizes is selected, otherwise the first one is used"
	paramVMSizeDefault   = "Standard_D8as_v5"

	paramSystemPoolVMSize     = "system-pool-vmsize"
	paramSystemPoolVMSizeDesc = "VMSize to be used on the system pool. The spot search only takes the user pool size into account, so the system pool size may not be available in the selected location"
	paramWorkersDesired       = "workers-desired"
	paramWorkersDesiredDesc   = "number of nodes on the user pool"
)

func GetAKSCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdAKS,
		Short: cmdAKSDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCreateAKS(), getDestroyAKS())
	return c
}

func getCreateAKS() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// viper splits a comma separated string slice on whitespace, so the
			// flag is read directly to keep the list of VM sizes intact.
			vmSizes, err := cmd.Flags().GetStringSlice(paramVMSize)
			if err != nil {
				return err
			}
			return azureAKS.Create(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&azureAKS.AKSArgs{
					Spot:              params.SpotArgs(),
					Location:          viper.GetString(paramLocation),
					KubernetesVersion: viper.GetString(paramVersion),
					OnlySystemPool:    viper.IsSet(paramOnlySystemPool),
					EnableAppRouting:  viper.IsSet(paramEnableAppRouting),
					VMSizes:           vmSizes,
					SystemPoolVMSize:  viper.GetString(paramSystemPoolVMSize),
					WorkersDesired:    viper.GetInt(paramWorkersDesired)})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramLocation, "", paramLocationDefault, paramLocationDesc)
	flagSet.StringSliceP(paramVMSize, "", []string{paramVMSizeDefault}, paramVMSizeDesc)
	flagSet.StringP(paramSystemPoolVMSize, "", azureAKS.DefaultSystemPoolVMSize, paramSystemPoolVMSizeDesc)
	flagSet.IntP(paramWorkersDesired, "", azureAKS.DefaultWorkersDesired, paramWorkersDesiredDesc)
	flagSet.StringP(paramVersion, "", defaultVersion, paramVersionDesc)
	flagSet.Bool(paramOnlySystemPool, false, paramOnlySystemPoolDesc)
	flagSet.Bool(paramEnableAppRouting, false, paramEnableAppRoutingDesc)
	params.AddSpotFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroyAKS() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return azureAKS.Destroy(&maptContext.ContextArgs{
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
