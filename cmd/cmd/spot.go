package cmd

import (
	spotPrice "github.com/adrianriobo/qenvs/pkg/infra/aws/modules/spot-price"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/util"
)

const (
	spotCmdName        string = "spot"
	spotCmdDescription string = "spot"

	instanceTypes      string = "instance-types"
	productDescription string = "product-description"
)

func init() {
	rootCmd.AddCommand(spotCmd)
	flagSet := pflag.NewFlagSet(spotCmdName, pflag.ExitOnError)
	flagSet.StringP(availabilityZones, "", "", availabilityZonesDesc)
	flagSet.StringP(instanceTypes, "i", "", "List of comma separated instace types")
	flagSet.StringP(productDescription, "p", "", "Filter instances by product description")
	spotCmd.Flags().AddFlagSet(flagSet)
}

var spotCmd = &cobra.Command{
	Use:   spotCmdName,
	Short: spotCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if _, err := spotPrice.BestSpotPriceInfo(
			util.SplitString(viper.GetString(availabilityZones), ","),
			util.SplitString(viper.GetString(instanceTypes), ","),
			viper.GetString(productDescription)); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
