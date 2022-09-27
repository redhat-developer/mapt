package cmd

import (
	"github.com/adrianriobo/qenvs/pkg/orchestrator"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/util"
)

const (
	spotCmdName        string = "spot"
	spotCmdDescription string = "spot"

	availabilityZones  string = "availability-zones"
	instanceTypes      string = "instance-types"
	productDescription string = "product-description"
)

func init() {
	rootCmd.AddCommand(spotCmd)
	flagSet := pflag.NewFlagSet(spotCmdName, pflag.ExitOnError)
	flagSet.StringP(availabilityZones, "a", "", "List of comma separated azs to check. If empty all will be searched")
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
		exec()
		return nil
	},
}

func exec() {
	if err := orchestrator.GetBestBidForSpot(
		util.SplitString(viper.GetString(availabilityZones), ","),
		util.SplitString(viper.GetString(instanceTypes), ","),
		viper.GetString(productDescription)); err != nil {
		logging.Error(err)
	}
}
