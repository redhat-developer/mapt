package hosts

import (
	params "github.com/adrianriobo/qenvs/cmd/cmd/constants"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/action/custom"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/action/fedora"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdCustom     = "fedora"
	cmdCustomDesc = "manage "

	amiID                     string = "ami"
	amiIDDesc                 string = "ID for the custom ami"
	instanceType              string = "instance-type"
	instanceTypeDesc          string = "type of instance"
	productDescription        string = "product-description"
	productDescriptionDesc    string = "Product description for the custom AMI: (Linux/UNIX, Windows or Red Hat Enterprise Linux)"
	productDescriptionDefault string = "Linux/UNIX"
)

func GetCustomCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdCustom,
		Short: cmdCustomDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(getCustomCreate(), getCustomDestroy())
	return c
}

func getCustomCreate() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			// Initialize context
			qenvsContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags))

			// Run create
			if err := custom.Create(
				&custom.Request{
					Prefix:             "main",
					AMI:                viper.GetString(amiID),
					InstanceType:       viper.GetString(instanceType),
					ProductDescription: viper.GetString(productDescription),
					Spot:               viper.IsSet(spot),
					Airgap:             viper.IsSet(airgap)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(amiID, "", "", amiIDDesc)
	flagSet.StringP(instanceType, "", "", instanceTypeDesc)
	flagSet.StringP(productDescription, "", productDescriptionDefault, productDescriptionDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	flagSet.Bool(spot, false, spotDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(amiID)
	if err != nil {
		logging.Error(err)
	}
	err = c.MarkPersistentFlagRequired(instanceType)
	if err != nil {
		logging.Error(err)
	}
	return c
}

func getCustomDestroy() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			qenvsContext.InitBase(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL))

			if err := fedora.Destroy(); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	return c
}
