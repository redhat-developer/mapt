package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/environment"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	hostCmdDestroyDescription string = "destroy host backed by project"
)

func init() {
	hostCmd.AddCommand(hostDestroyCmd)
	flagSet := pflag.NewFlagSet(destroyCmdName, pflag.ExitOnError)
	flagSet.StringP(projectName, "", "", projectNameDesc)
	flagSet.StringP(backedURL, "", "", backedURLDesc)
	hostDestroyCmd.Flags().AddFlagSet(flagSet)
}

var hostDestroyCmd = &cobra.Command{
	Use:   destroyCmdName,
	Short: hostCmdDestroyDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := environment.Destroy(
			"qenvs",
			"file:///tmp/qenvs"); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
