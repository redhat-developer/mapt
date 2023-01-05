package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	amireplication "github.com/adrianriobo/qenvs/pkg/infra/aws/modules/ami-replication"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	amiReplicasCreateCmdDescription string = "create replicas on all regions based on a base source image"
)

func init() {
	amiCmd.AddCommand(amiReplicasCreateCmd)
	flagSet := pflag.NewFlagSet(createCmdName, pflag.ExitOnError)
	flagSet.StringP(projectName, "", "", projectNameDesc)
	flagSet.StringP(backedURL, "", "", backedURLDesc)
	flagSet.StringP(amiIDName, "", "", amiIDDesc)
	flagSet.StringP(amiNameName, "", "", amiNameDesc)
	flagSet.StringP(amiSourceRegion, "", "", amiSourceRegionDesc)
	amiReplicasCreateCmd.Flags().AddFlagSet(flagSet)
}

var amiReplicasCreateCmd = &cobra.Command{
	Use:   createCmdName,
	Short: amiReplicasCreateCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := amireplication.CreateReplicas(
			viper.GetString(projectName),
			viper.GetString(backedURL),
			viper.GetString(amiIDName),
			viper.GetString(amiNameName),
			viper.GetString(amiSourceRegion)); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
