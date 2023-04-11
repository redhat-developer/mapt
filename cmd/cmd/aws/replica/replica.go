package replica

import (
	params "github.com/adrianriobo/qenvs/cmd/cmd/constants"
	amireplication "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/ami-replication"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmd     = "ami-replica"
	cmdDesc = "replica operations"
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
	c.AddCommand(getCreate(), getDestroy())
	return c
}

func getCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := amireplication.CreateReplicas(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.AMIIDName),
				viper.GetString(params.AMINameName),
				viper.GetString(params.AMISourceRegion)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.AMIIDName, "", "", params.AMIIDDesc)
	flagSet.StringP(params.AMINameName, "", "", params.AMINameDesc)
	flagSet.StringP(params.AMISourceRegion, "", "", params.AMISourceRegionDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroy() *cobra.Command {
	return &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := amireplication.DestroyReplicas(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
}
