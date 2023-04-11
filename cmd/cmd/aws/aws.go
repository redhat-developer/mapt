package aws

import (
	"github.com/adrianriobo/qenvs/cmd/cmd/aws/host"
	"github.com/adrianriobo/qenvs/cmd/cmd/aws/replica"
	"github.com/adrianriobo/qenvs/cmd/cmd/aws/spot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cmd     = "aws"
	cmdDesc = "aws operations"
)

func GetCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmd,
		Short: cmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}

	c.AddCommand(
		replica.GetCmd(),
		spot.GetCmd(),
		host.GetCmd())
	return c
}
