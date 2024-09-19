package azure

import (
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/hosts"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/azure/services"
	"github.com/spf13/cobra"
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

	c.AddCommand(
		hosts.GetWindowsDesktopCmd(),
		hosts.GetUbuntuCmd(),
		hosts.GetRHELCmd(),
		services.GetAKSCmd())
	return c
}
