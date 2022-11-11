package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	amiCmdName        string = "ami-replica"
	amiCmdDescription string = "manage ami for supported hosts"
)

func init() {
	rootCmd.AddCommand(amiCmd)
	// flagSet := pflag.NewFlagSet(amiCmdName, pflag.ExitOnError)
	// amiCmd.Flags().AddFlagSet(flagSet)
}

var amiCmd = &cobra.Command{
	Use:   amiCmdName,
	Short: amiCmdDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		return nil
	},
}
