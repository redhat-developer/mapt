package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/environment"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	corpDestroyCmdName        string = "destroy"
	corpCmdDestroyDescription string = "destroy MS corporate environment"
)

func init() {
	corpCmd.AddCommand(corpMicrosoftDestroyCmd)
}

var corpMicrosoftDestroyCmd = &cobra.Command{
	Use:   corpDestroyCmdName,
	Short: corpCmdDestroyDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := environment.Destroy("qenvs", "file:///tmp/qenvs"); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
