package azure

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/hosts"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/services"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmd     = "azure"
	cmdDesc = "azure operations"
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
		hosts.GetWindowsDesktopCmd(),
		hosts.GetUbuntuCmd(),
		hosts.GetRHELCmd(),
		hosts.GetRHELAICmd(),
		hosts.GetFedoraCmd(),
		services.GetAKSCmd(),
		services.GetKindCmd())
	return c
}
