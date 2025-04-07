package aws

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/hosts"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/services"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	
	flagSet := pflag.NewFlagSet(cmd, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)

	c.AddCommand(
		hosts.GetMacCmd(),
		hosts.GetWindowsCmd(),
		hosts.GetRHELCmd(),
		hosts.GetFedoraCmd(),
		services.GetMacPoolCmd(),
		services.GetEKSCmd())
	return c
}
