package aws

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/hosts"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/services"
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
		hosts.GetMacCmd(),
		hosts.GetWindowsCmd(),
		hosts.GetRHELCmd(),
		hosts.GetFedoraCmd(),
		services.GetMacPoolCmd())
	return c
}
