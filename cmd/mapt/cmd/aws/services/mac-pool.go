package services

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/params"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	macpool "github.com/redhat-developer/mapt/pkg/provider/aws/action/mac-pool"
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
		createMP(),
		destroyMP(),
		houseKeep(),
		request(),
		release())
	return c
}

func createMP() *cobra.Command {
	c := &cobra.Command{
		Use:   params.CreateCmdName,
		Short: params.CreateCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return macpool.Create(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&macpool.MacPoolRequestArgs{
					Prefix:          "main",
					PoolName:        viper.GetString(paramName),
					Architecture:    viper.GetString(awsParams.MACArch),
					OSVersion:       viper.GetString(awsParams.MACOSVersion),
					OfferedCapacity: viper.GetInt(paramOfferedCapacity),
					MaxSize:         viper.GetInt(paramMaxSize),
					FixedLocation:   viper.IsSet(awsParams.MACFixedLocation)})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringP(paramName, "", "", paramNameDesc)
	flagSet.Int(paramOfferedCapacity, paramOfferedCapacityDefault, paramOfferedCapacityDesc)
	flagSet.Int(paramMaxSize, paramMaxSizeDefault, paramMaxSizeDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersionDefault, awsParams.MACOSVersionDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.Bool(awsParams.MACFixedLocation, false, awsParams.MACFixedLocationDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func destroyMP() *cobra.Command {
	c := &cobra.Command{
		Use:   params.DestroyCmdName,
		Short: params.DestroyCmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return macpool.Destroy(&maptContext.ContextArgs{
				Context:     cmd.Context(),
				ProjectName: viper.GetString(params.ProjectName),
				BackedURL:   viper.GetString(params.BackedURL),
				Debug:       viper.IsSet(params.Debug),
				DebugLevel:  viper.GetUint(params.DebugLevel),
				KeepState:   viper.IsSet(params.KeepState),
			})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	flagSet.Bool(params.KeepState, false, params.KeepStateDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func houseKeep() *cobra.Command {
	c := &cobra.Command{
		Use:   cmdHousekeep,
		Short: cmdHousekeepDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return macpool.HouseKeeper(
				&maptContext.ContextArgs{
					Context:     cmd.Context(),
					ProjectName: viper.GetString(params.ProjectName),
					BackedURL:   viper.GetString(params.BackedURL),
					Serverless:  viper.IsSet(params.Serverless),
					Debug:       viper.IsSet(params.Debug),
					DebugLevel:  viper.GetUint(params.DebugLevel),
					Tags:        viper.GetStringMapString(params.Tags),
				},
				&macpool.MacPoolRequestArgs{
					Prefix:          "main",
					PoolName:        viper.GetString(paramName),
					Architecture:    viper.GetString(awsParams.MACArch),
					OSVersion:       viper.GetString(awsParams.MACOSVersion),
					OfferedCapacity: viper.GetInt(paramOfferedCapacity),
					MaxSize:         viper.GetInt(paramMaxSize),
					FixedLocation:   viper.IsSet(awsParams.MACFixedLocation)})
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
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

func request() *cobra.Command {
	c := &cobra.Command{
		Use:   awsParams.MACRequestCmd,
		Short: awsParams.MACRequestCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return macpool.Request(
				&maptContext.ContextArgs{
					Context:       cmd.Context(),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					CirrusPWArgs:  params.CirrusPersistentWorkerArgs(),
					GHRunnerArgs:  params.GithubRunnerArgs(),
					GLRunnerArgs:  params.GitLabRunnerArgs(),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&macpool.RequestMachineArgs{
					PoolName:     viper.GetString(paramName),
					Architecture: viper.GetString(awsParams.MACArch),
					OSVersion:    viper.GetString(awsParams.MACOSVersion),
					Timeout:      viper.GetString(params.Timeout),
				})
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACRequestCmd, pflag.ExitOnError)
	flagSet.StringP(params.ConnectionDetailsOutput, "", "", params.ConnectionDetailsOutputDesc)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramName, "", "", paramNameDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	params.AddGHActionsFlags(flagSet)
	params.AddCirrusFlags(flagSet)
	params.AddGitLabRunnerFlags(flagSet)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}

func release() *cobra.Command {
	c := &cobra.Command{
		Use:   awsParams.MACReleaseCmd,
		Short: awsParams.MACReleaseCmd,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			return macpool.Release(
				&maptContext.ContextArgs{
					Context:    cmd.Context(),
					Debug:      viper.IsSet(params.Debug),
					DebugLevel: viper.GetUint(params.DebugLevel),
					Serverless: viper.IsSet(params.Serverless),
				},
				viper.GetString(awsParams.MACDHID))
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACReleaseCmd, pflag.ExitOnError)
	flagSet.StringP(awsParams.MACDHID, "", "", awsParams.MACDHIDDesc)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	err := c.MarkPersistentFlagRequired(awsParams.MACDHID)
	if err != nil {
		logging.Error(err)
	}
	return c
}
