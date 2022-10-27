package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	hostCmdName        string = "host"
	hostCmdDescription string = "manage supported hosts"
)

func init() {
	rootCmd.AddCommand(hostCmd)
}

var hostCmd = &cobra.Command{
	Use:   hostCmdName,
	Short: hostCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		return nil
	},
}
