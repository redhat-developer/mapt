package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/vpc/stacks"
	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	corpMicrosoftCreateCmdName        string = "create"
	corpMicrosoftCmdCreateDescription string = "create MS corporate environment"
)

func init() {
	corpMicrosoftCmd.AddCommand(corpMicrosoftCreateCmd)
	flagSet := pflag.NewFlagSet(corpMicrosoftCmdName, pflag.ExitOnError)
	flagSet.StringP(availabilityZones, "", "", availabilityZonesDesc)
	flagSet.StringP(cidr, "", "", cidrDesc)
	flagSet.StringP(publicSubnetCIDRs, "", "", publicSubnetCIDRsDesc)
	flagSet.StringP(privateSubnetCIDRs, "", "", privateSubnetCIDRsDesc)
	corpMicrosoftCreateCmd.Flags().AddFlagSet(flagSet)
}

var corpMicrosoftCreateCmd = &cobra.Command{
	Use:   corpMicrosoftCreateCmdName,
	Short: corpMicrosoftCmdCreateDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := stacks.CreateVPC(
			"qenvs", "file:///tmp/qenvs",
			viper.GetString(cidr),
			util.SplitString(viper.GetString(availabilityZones), ","),
			util.SplitString(viper.GetString(privateSubnetCIDRs), ","),
			util.SplitString(viper.GetString(publicSubnetCIDRs), ",")); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
