package cmd

import (
	spotPrice "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/spot-price"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	spotCmdName        string = "spot"
	spotCmdDescription string = "spot price prediction"
)

func init() {
	rootCmd.AddCommand(spotCmd)
	flagSet := pflag.NewFlagSet(spotCmdName, pflag.ExitOnError)
	flagSet.StringP(availabilityZones, "", "", availabilityZonesDesc)
	flagSet.StringP(supportedHostID, "", "", supportedHostIDDesc)
	spotCmd.Flags().AddFlagSet(flagSet)
}

var spotCmd = &cobra.Command{
	Use:   spotCmdName,
	Short: spotCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		// util.SplitString(viper.GetString(availabilityZones), ","),
		if _, err := spotPrice.BestSpotPriceInfo(
			viper.GetString(supportedHostID)); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
