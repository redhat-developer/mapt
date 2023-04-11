package spot

import (
	params "github.com/adrianriobo/qenvs/cmd/cmd/constants"
	spotPrice "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/spot-price"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmd     = "spot"
	cmdDesc = "spot operations"
)

func GetCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmd,
		Short: cmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if _, err := spotPrice.BestSpotPriceInfo(
				viper.GetString(params.SupportedHostID)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(cmd, pflag.ExitOnError)
	flagSet.StringP(params.SupportedHostID, "", "", params.SupportedHostIDDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
