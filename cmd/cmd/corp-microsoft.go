package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	corpMicrosoftCmdName        string = "corp"
	corpMicrosoftCmdDescription string = "corp"

	cidr                   string = "network-cidr"
	cidrDesc               string = "cidr block for network"
	publicSubnetCIDRs      string = "public-subnet-cidrs"
	publicSubnetCIDRsDesc  string = "List of comma separated cidrs per public subnet."
	privateSubnetCIDRs     string = "private-subnet-cidrs"
	privateSubnetCIDRsDesc string = "List of comma separated cidrs per private subnet."
)

func init() {
	rootCmd.AddCommand(corpMicrosoftCmd)
}

var corpMicrosoftCmd = &cobra.Command{
	Use:   corpMicrosoftCmdName,
	Short: corpMicrosoftCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		// if err := orchestrator.GetBestBidForSpot(
		// 	util.SplitString(viper.GetString(availabilityZones), ","),
		// 	util.SplitString(viper.GetString(instanceTypes), ","),
		// 	viper.GetString(productDescription)); err != nil {
		// 	logging.Error(err)
		// }
		return nil
	},
}
