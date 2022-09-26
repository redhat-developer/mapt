package cmd

import (
	"strings"

	"github.com/adrianriobo/qenvs/pkg/infra/orchestrator"
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
	instanceType       string = "instance-type"
	productDescription string = "product-description"
)

func init() {
	rootCmd.AddCommand(spotCmd)
	flagSet := pflag.NewFlagSet(spotCmdName, pflag.ExitOnError)
	flagSet.StringP(availabilityZones, "a", "", "List of comma separated azs to check. If empty all will be searched")
	flagSet.StringP(instanceType, "i", "", "Instace type")
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
		parsetAvailabilityZones(viper.GetString(availabilityZones)),
		viper.GetString(instanceType),
		viper.GetString(productDescription)); err != nil {
		logging.Error(err)
	}
}

func parsetAvailabilityZones(availabilityZones string) []string {
	return util.If(strings.Contains(availabilityZones, ","),
		strings.Split(viper.GetString(availabilityZones), ","),
		[]string{})
}
