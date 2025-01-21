package services

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	macpool "github.com/redhat-developer/mapt/pkg/provider/aws/action/mac-pool"
	"github.com/redhat-developer/mapt/pkg/util/ghactions"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cmdMacPool     = "mac-pool"
	cmdMacPoolDesc = "mac pool operations"

	cmdHousekeep     = "house-keep"
	cmdHousekeepDesc = "house keeping for mac pool. Detroy old machines on over capacity and create new ones if capacity not meet"

	paramName                   = "name"
	paramNameDesc               = "pool name it is a unique identifier for the pool. The name should be unique for the whole AWS account"
	paramOfferedCapacity        = "offered-capacity"
	paramOfferedCapacityDesc    = "offered capacity to accept new workloads at any given time. Limited by max pool size"
	paramOfferedCapacityDefault = 1
	paramMaxSize                = "max-size"
	paramMaxSizeDesc            = "max number of machines in the pool"
	paramMaxSizeDefault         = 2
)

func GetMacPoolCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdMacPool,
		Short: cmdMacPoolDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return nil
		},
	}
	c.AddCommand(
		getCreateMacPool(),
		getHouseKeepingMacPool(),
		getRequest(),
		getRelease())
	return c
}

func getCreateMacPool() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize context
			maptContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel),
				false)

			if err := macpool.Create(
				&macpool.RequestArgs{
					Prefix:          "main",
					PoolName:        viper.GetString(paramName),
					Architecture:    viper.GetString(awsParams.MACArch),
					OSVersion:       viper.GetString(awsParams.MACOSVersion),
					OfferedCapacity: viper.GetInt(paramOfferedCapacity),
					MaxSize:         viper.GetInt(paramMaxSize),
					FixedLocation:   viper.IsSet(awsParams.MACFixedLocation)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramName, "", "", paramNameDesc)
	flagSet.Int(paramOfferedCapacity, paramOfferedCapacityDefault, paramOfferedCapacityDesc)
	flagSet.Int(paramMaxSize, paramMaxSizeDefault, paramMaxSizeDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.Bool(awsParams.MACFixedLocation, false, awsParams.MACFixedLocationDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getHouseKeepingMacPool() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdHousekeep,
		Short: cmdHousekeepDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize context
			maptContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel),
				viper.IsSet(params.Serverless))

			if err := macpool.HouseKeeper(
				&macpool.RequestArgs{
					Prefix:          "main",
					PoolName:        viper.GetString(paramName),
					Architecture:    viper.GetString(awsParams.MACArch),
					OSVersion:       viper.GetString(awsParams.MACOSVersion),
					OfferedCapacity: viper.GetInt(paramOfferedCapacity),
					MaxSize:         viper.GetInt(paramMaxSize),
					FixedLocation:   viper.IsSet(awsParams.MACFixedLocation)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramName, "", "", paramNameDesc)
	flagSet.Int(paramOfferedCapacity, paramOfferedCapacityDefault, paramOfferedCapacityDesc)
	flagSet.Int(paramMaxSize, paramMaxSizeDefault, paramMaxSizeDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.Bool(awsParams.MACFixedLocation, false, awsParams.MACFixedLocationDesc)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getRequest() *cobra.Command {
	c := &cobra.Command{
		Use:   awsParams.MACRequestCmd,
		Short: awsParams.MACRequestCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize gh actions runner if needed
			if viper.IsSet(params.InstallGHActionsRunner) {
				err := ghactions.InitGHRunnerArgs(
					viper.GetString(params.GHActionsRunnerToken),
					viper.GetString(params.GHActionsRunnerName),
					viper.GetString(params.GHActionsRunnerRepo),
					viper.GetStringSlice(params.GHActionsRunnerLabels))
				if err != nil {
					logging.Fatal(err)
				}
			}

			// Initialize context
			maptContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel),
				viper.IsSet(params.Serverless))

			if err := macpool.Request(
				&macpool.RequestMachineArgs{
					PoolName:             viper.GetString(paramName),
					Architecture:         viper.GetString(awsParams.MACArch),
					OSVersion:            viper.GetString(awsParams.MACOSVersion),
					SetupGHActionsRunner: viper.IsSet(params.InstallGHActionsRunner)}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACRequestCmd, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramName, "", "", paramNameDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func getRelease() *cobra.Command {
	c := &cobra.Command{
		Use:   awsParams.MACReleaseCmd,
		Short: awsParams.MACReleaseCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize context
			maptContext.Init(
				viper.GetString(params.ProjectName),
				viper.GetString(params.BackedURL),
				viper.GetString(params.ConnectionDetailsOutput),
				viper.GetStringMapString(params.Tags),
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel),
				viper.IsSet(params.Serverless))

			if err := macpool.Release(
				&macpool.ReleaseMachineArgs{
					MachineID: viper.GetString(awsParams.MACDHID)},
				viper.IsSet(params.Debug),
				viper.GetUint(params.DebugLevel)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACReleaseCmd, pflag.ExitOnError)
	flagSet.StringP(awsParams.MACDHID, "", "", awsParams.MACDHIDDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(awsParams.MACDHID)
	if err != nil {
		logging.Error(err)
	}
	return c
}
