package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/adrianriobo/qenvs/pkg/infra/aws/vpc/stacks"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	corpMicrosoftDestroyCmdName        string = "destroy"
	corpMicrosoftCmdDestroyDescription string = "destrioy MS corporate environment"
)

func init() {
	corpMicrosoftCmd.AddCommand(corpMicrosoftDestroyCmd)
}

var corpMicrosoftDestroyCmd = &cobra.Command{
	Use:   corpMicrosoftDestroyCmdName,
	Short: corpMicrosoftCmdDestroyDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := stacks.DestroyVPC("qenvs", "file:///tmp/qenvs"); err != nil {
			logging.Error(err)
		}
		return nil
	},
}
