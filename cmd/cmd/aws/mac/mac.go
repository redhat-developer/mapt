package mac

import (
	"fmt"

	params "github.com/adrianriobo/qenvs/cmd/cmd/constants"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/action/mac"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmd               = "mac"
	cmdDesc           = "create mac instances"
	checkStateCmd     = "check-state"
	checkStateCmdDesc = "check the state for a dedicated mac machine"

	arch              string = "arch"
	archDesc          string = "mac architecture allowed values x86, m1, m2"
	archDefault       string = "m2"
	osVersion         string = "version"
	osVersionDesc     string = "macos operating system vestion 11, 12 on x86 and m1; 13, 14 on all archs"
	osDefault         string = "14"
	hostID            string = "host-id"
	hostIDDesc        string = "host id to create the mac instance. If the param is not pass the dedicated host will be created"
	onlyHost          string = "only-host"
	onlyHostDesc      string = "if this flag is set only the host will be created / destroyed"
	onlyMachine       string = "only-machine"
	onlyMachineDesc   string = "if this flag is set only the machine will be destroyed"
	fixedLocation     string = "fixed-location"
	fixedLocationDesc string = "if this flag is set the host will be created only on the region set by the AWS Env (AWS_DEFAULT_REGION)"
	airgap            string = "airgap"
	airgapDesc        string = "if this flag is set the host will be created as airgap machine. Access will done through a bastion"
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
	c.AddCommand(getCreate(), getDestroy(), getCheckState())
	return c
}

func getCheckState() *cobra.Command {
	c := &cobra.Command{
		Use:   checkStateCmd,
		Short: checkStateCmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			state, err := mac.CheckState(viper.GetString(hostID))
			if err != nil {
				logging.Error(err)
				return err
			}
			fmt.Printf("%s", *state)
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(hostID, "", "", hostIDDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	if err := c.MarkPersistentFlagRequired(hostID); err != nil {
		logging.Error(err)
	}
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

			// Initialize context
			qenvsContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags))

			// Run create
			if err := mac.Create(
				&mac.MacRequest{
					Prefix:        "main",
					Architecture:  viper.GetString(arch),
					Version:       viper.GetString(osVersion),
					HostID:        viper.GetString(hostID),
					OnlyHost:      viper.IsSet(onlyHost),
					OnlyMachine:   viper.IsSet(onlyMachine),
					FixedLocation: viper.IsSet(fixedLocation),
					Airgap:        viper.IsSet(airgap)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(arch, "", archDefault, archDesc)
	flagSet.StringP(osVersion, "", osDefault, osVersionDesc)
	flagSet.StringP(hostID, "", "", hostIDDesc)
	flagSet.Bool(onlyHost, false, onlyHostDesc)
	flagSet.Bool(onlyMachine, false, onlyMachineDesc)
	flagSet.Bool(fixedLocation, false, fixedLocationDesc)
	flagSet.Bool(airgap, false, airgapDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getDestroy() *cobra.Command {
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

			if err := mac.Destroy(
				&mac.MacRequest{
					Prefix:      "main",
					OnlyHost:    viper.IsSet(onlyHost),
					OnlyMachine: viper.IsSet(onlyMachine)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.Bool(onlyHost, false, onlyHostDesc)
	flagSet.Bool(onlyMachine, false, onlyMachineDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
