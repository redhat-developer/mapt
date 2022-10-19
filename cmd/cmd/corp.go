package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	corpCmdName        string = "corp"
	corpCmdDescription string = "corp"
)

func init() {
	rootCmd.AddCommand(corpCmd)
}

var corpCmd = &cobra.Command{
	Use:   corpCmdName,
	Short: corpCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		return nil
	},
}
