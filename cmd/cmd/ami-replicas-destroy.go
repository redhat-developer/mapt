package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	amireplication "github.com/adrianriobo/qenvs/pkg/infra/aws/modules/ami-replication"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	amiReplicateDestroyCmdDescription string = "destroy replicas on all regions based on a base source image"
)

func init() {
	amiCmd.AddCommand(amiReplicasDestroyCmd)
	flagSet := pflag.NewFlagSet(destroyCmdName, pflag.ExitOnError)
	flagSet.StringP(projectName, "", "", projectNameDesc)
	flagSet.StringP(backedURL, "", "", backedURLDesc)
	amiReplicasDestroyCmd.Flags().AddFlagSet(flagSet)
}

var amiReplicasDestroyCmd = &cobra.Command{
	Use:   destroyCmdName,
	Short: amiReplicateDestroyCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := amireplication.DestroyReplicas(
			viper.GetString(projectName),
			viper.GetString(backedURL)); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
