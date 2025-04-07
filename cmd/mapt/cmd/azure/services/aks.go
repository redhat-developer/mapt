package services

import (
	"fmt"

	azparams "github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	azureAKS "github.com/redhat-developer/mapt/pkg/provider/azure/action/aks"
	spotAzure "github.com/redhat-developer/mapt/pkg/spot/azure"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdAKS     = "aks"
	cmdAKSDesc = "aks operations"

	paramVersion              = "version"
	paramVersionDesc          = "AKS K8s cluster version"
	paramVMSizeDesc           = "VMSize to be used on the user pool. Typically this is used to provision spot node pools"
	defaultVersion            = "1.31"
	paramOnlySystemPool       = "only-system-pool"
	paramOnlySystemPoolDesc   = "if we do not need bunch of resources we can run only the systempool. More info https://learn.microsoft.com/es-es/azure/aks/use-system-pools?tabs=azure-cli#system-and-user-node-pools"
	paramEnableAppRouting     = "enable-app-routing"
	paramEnableAppRoutingDesc = "enable application routing add-on with NGINX"
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

			// ParseEvictionRate
			var spotToleranceValue = spotAzure.DefaultEvictionRate
			if viper.IsSet(azparams.ParamSpotTolerance) {
				var ok bool
				spotToleranceValue, ok = spotAzure.ParseEvictionRate(
					viper.GetString(azparams.ParamSpotTolerance))
				if !ok {
					return fmt.Errorf("%s is not a valid spot tolerance value", viper.GetString(azparams.ParamSpotTolerance))
				}
			}

			if err := azureAKS.Create(
				&maptContext.ContextArgs{
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&azureAKS.AKSRequest{
					Prefix:              viper.GetString(params.ProjectName),
					Location:            viper.GetString(azparams.ParamLocation),
					KubernetesVersion:   viper.GetString(paramVersion),
					OnlySystemPool:      viper.IsSet(paramOnlySystemPool),
					EnableAppRouting:    viper.IsSet(paramEnableAppRouting),
					VMSize:              viper.GetString(azparams.ParamVMSize),
					Spot:                viper.IsSet(azparams.ParamSpot),
					SpotTolerance:       spotToleranceValue,
					SpotExcludedRegions: viper.GetStringSlice(azparams.ParamSpotExcludedRegions)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(azparams.ParamLocation, "", azparams.DefaultLocation, azparams.ParamLocationDesc)
	flagSet.StringP(azparams.ParamVMSize, "", azparams.DefaultVMSize, paramVMSizeDesc)
	flagSet.StringP(paramVersion, "", defaultVersion, paramVersionDesc)
	flagSet.Bool(azparams.ParamSpot, false, azparams.ParamSpotDesc)
	flagSet.Bool(paramOnlySystemPool, false, paramOnlySystemPoolDesc)
	flagSet.Bool(paramEnableAppRouting, false, paramEnableAppRoutingDesc)
	flagSet.StringP(azparams.ParamSpotTolerance, "", azparams.DefaultSpotTolerance, azparams.ParamSpotToleranceDesc)
	flagSet.StringSliceP(azparams.ParamSpotExcludedRegions, "", []string{}, azparams.ParamSpotExcludedRegionsDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroyAKS() *cobra.Command {
	return &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := azureAKS.Destroy(
				&maptContext.ContextArgs{
					ProjectName: viper.GetString(params.ProjectName),
					BackedURL:   viper.GetString(params.BackedURL),
					Debug:       viper.IsSet(params.Debug),
					DebugLevel:  viper.GetUint(params.DebugLevel),
				}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
