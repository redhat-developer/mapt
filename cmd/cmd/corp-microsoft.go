package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	corpMicrosoftCmdName        string = "corp"
	corpMicrosoftCmdDescription string = "corp"
)

func init() {
	rootCmd.AddCommand(corpMicrosoftCmd)
}

var corpMicrosoftCmd = &cobra.Command{
	Use:   corpMicrosoftCmdName,
	Short: corpMicrosoftCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		return nil
	},
}
