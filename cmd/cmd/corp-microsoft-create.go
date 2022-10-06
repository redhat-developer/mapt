package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/vpc"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	corpMicrosoftCreateCmdName        string = "create"
	corpMicrosoftCmdCreateDescription string = "create MS corporate environment"

	cidr                   string = "network-cidr"
	cidrDesc               string = "cidr block for network"
	publicSubnetCIDRs      string = "public-subnet-cidrs"
	publicSubnetCIDRsDesc  string = "List of comma separated cidrs per public subnet."
	privateSubnetCIDRs     string = "private-subnet-cidrs"
	privateSubnetCIDRsDesc string = "List of comma separated cidrs per private subnet. Category 2 or 3 according to https://www.rfc-editor.org/rfc/rfc1918"
	intraSubnetCIDRs       string = "intra-subnet-cidrs"
	intraSubnetCIDRsDesc   string = "List of comma separated cidrs per private subnet.Category 1 according to https://www.rfc-editor.org/rfc/rfc1918"
)

func init() {
	corpMicrosoftCmd.AddCommand(corpMicrosoftCreateCmd)
	flagSet := pflag.NewFlagSet(corpMicrosoftCmdName, pflag.ExitOnError)
	flagSet.StringP(availabilityZones, "", "", availabilityZonesDesc)
	flagSet.StringP(cidr, "", "", cidrDesc)
	flagSet.StringP(publicSubnetCIDRs, "", "", publicSubnetCIDRsDesc)
	flagSet.StringP(privateSubnetCIDRs, "", "", privateSubnetCIDRsDesc)
	flagSet.StringP(intraSubnetCIDRs, "", "", intraSubnetCIDRsDesc)
	corpMicrosoftCreateCmd.Flags().AddFlagSet(flagSet)
}

var corpMicrosoftCreateCmd = &cobra.Command{
	Use:   corpMicrosoftCreateCmdName,
	Short: corpMicrosoftCmdCreateDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := vpc.CreateNetwork(
			"qenvs", "file:///tmp/qenvs",
			viper.GetString(cidr),
			util.SplitString(viper.GetString(availabilityZones), ","),
			util.SplitString(viper.GetString(publicSubnetCIDRs), ","),
			util.SplitString(viper.GetString(privateSubnetCIDRs), ","),
			util.SplitString(viper.GetString(intraSubnetCIDRs), ",")); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
