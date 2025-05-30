package services

import (
	awsParams "github.com/redhat-developer/mapt/cmd/mapt/cmd/aws/constants"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
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

	// Pool params
	paramName                   = "name"
	paramNameDesc               = "pool name it is a unique identifier for the pool. The name should be unique for the whole AWS account"
	paramOfferedCapacity        = "offered-capacity"
	paramOfferedCapacityDesc    = "offered capacity to accept new workloads at any given time. Limited by max pool size"
	paramOfferedCapacityDefault = 1
	paramMaxSize                = "max-size"
	paramMaxSizeDesc            = "max number of machines in the pool"
	paramMaxSizeDefault         = 2
	// Machines in the pool params
	paramVPCID          = "vpcid"
	paramVPCIDDesc      = "VPC Id to setup mac machines"
	paramVPCIDDefault   = ""
	paramSSHSGID        = "ssh-sgid"
	paramSSHSGIDDesc    = "Security group Id to securize ssh access to machines. SSH can only be used from instances with this SG."
	paramSSHSGIDDefault = ""
	// Request / Release
	paramTicket        = "ticket"
	paramTicketDesc    = "this is a unique identifier to tag the dedicated host meanwhile it is being locked to identify the request which locked it. It will be used on release to identify the machine being released"
	paramTicketDefault = ""
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
			if err := macpool.Create(
				&maptContext.ContextArgs{
					ProjectName:   viper.GetString(params.ProjectName),
					BackedURL:     viper.GetString(params.BackedURL),
					ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
					Debug:         viper.IsSet(params.Debug),
					DebugLevel:    viper.GetUint(params.DebugLevel),
					Tags:          viper.GetStringMapString(params.Tags),
				},
				&macpool.PoolRequestArgs{
					Prefix:          "main",
					Name:            viper.GetString(paramName),
					Architecture:    viper.GetString(awsParams.MACArch),
					OSVersion:       viper.GetString(awsParams.MACOSVersion),
					OfferedCapacity: viper.GetInt(paramOfferedCapacity),
					MaxSize:         viper.GetInt(paramMaxSize)}); err != nil {
				logging.Error(err)
			}
			return nil
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

			if err := macpool.Destroy(&maptContext.ContextArgs{
				ProjectName: viper.GetString(params.ProjectName),
				BackedURL:   viper.GetString(params.BackedURL),
				Debug:       viper.IsSet(params.Debug),
				DebugLevel:  viper.GetUint(params.DebugLevel),
			}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
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

			if err := macpool.HouseKeeper(
				&maptContext.ContextArgs{
					ProjectName: viper.GetString(params.ProjectName),
					BackedURL:   viper.GetString(params.BackedURL),
					Serverless:  viper.IsSet(params.Serverless),
					Debug:       viper.IsSet(params.Debug),
					DebugLevel:  viper.GetUint(params.DebugLevel),
					Tags:        viper.GetStringMapString(params.Tags),
				},
				&macpool.HouseKeepRequestArgs{
					Pool: &macpool.PoolRequestArgs{
						Prefix:          "main",
						Name:            viper.GetString(paramName),
						Architecture:    viper.GetString(awsParams.MACArch),
						OSVersion:       viper.GetString(awsParams.MACOSVersion),
						OfferedCapacity: viper.GetInt(paramOfferedCapacity),
						MaxSize:         viper.GetInt(paramMaxSize),
					},
					Machine: &macpool.MachineRequestArgs{
						VPCID:   viper.GetString(paramVPCID),
						SSHSGID: viper.GetString(paramSSHSGID)},
				}); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(params.CreateCmdName, pflag.ExitOnError)
	params.AddCommonFlags(flagSet)
	flagSet.StringToStringP(params.Tags, "", nil, params.TagsDesc)
	flagSet.StringP(paramName, "", "", paramNameDesc)
	flagSet.Int(paramOfferedCapacity, paramOfferedCapacityDefault, paramOfferedCapacityDesc)
	flagSet.Int(paramMaxSize, paramMaxSizeDefault, paramMaxSizeDesc)
	flagSet.StringP(awsParams.MACArch, "", awsParams.MACArchDefault, awsParams.MACArchDesc)
	flagSet.StringP(awsParams.MACOSVersion, "", awsParams.MACOSVersion, awsParams.MACOSVersionDefault)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.StringP(paramVPCID, "", paramVPCIDDefault, paramVPCIDDesc)
	flagSet.StringP(paramSSHSGID, "", paramSSHSGIDDefault, paramSSHSGIDDesc)
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

			ctx := &maptContext.ContextArgs{
				ResultsOutput: viper.GetString(params.ConnectionDetailsOutput),
				Debug:         viper.IsSet(params.Debug),
				DebugLevel:    viper.GetUint(params.DebugLevel),
				Serverless:    viper.IsSet(params.Serverless),
				Remote:        viper.IsSet(params.Remote),
				Tags:          viper.GetStringMapString(params.Tags),
			}

			if viper.IsSet(params.GHActionsRunnerToken) {
				ctx.GHRunnerArgs = &github.GithubRunnerArgs{
					Token:    viper.GetString(params.GHActionsRunnerToken),
					RepoURL:  viper.GetString(params.GHActionsRunnerRepo),
					Labels:   viper.GetStringSlice(params.GHActionsRunnerLabels),
					Platform: &github.Linux,
					Arch: awsParams.MACArchAsGithubArch(
						viper.GetString(awsParams.MACArch)),
				}
			}

			if viper.IsSet(params.CirrusPWToken) {
				ctx.CirrusPWArgs = &cirrus.PersistentWorkerArgs{
					Token:    viper.GetString(params.CirrusPWToken),
					Labels:   viper.GetStringMapString(params.CirrusPWLabels),
					Platform: &cirrus.Darwin,
					Arch: awsParams.MACArchAsCirrusArch(
						viper.GetString(awsParams.MACArch)),
				}
			}

			if err := macpool.Request(
				ctx,
				&macpool.RequestMachineArgs{
					PoolName:     viper.GetString(paramName),
					Architecture: viper.GetString(awsParams.MACArch),
					OSVersion:    viper.GetString(awsParams.MACOSVersion),
					Machine: &macpool.MachineRequestArgs{
						VPCID:   viper.GetString(paramVPCID),
						SSHSGID: viper.GetString(paramSSHSGID),
					},
					Ticket:  viper.GetString(paramTicket),
					Timeout: viper.GetString(params.Timeout),
				}); err != nil {
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
	flagSet.StringP(params.Timeout, "", "", params.TimeoutDesc)
	flagSet.StringP(paramVPCID, "", paramVPCIDDefault, paramVPCIDDesc)
	flagSet.StringP(paramSSHSGID, "", paramSSHSGIDDefault, paramSSHSGIDDesc)
	flagSet.StringP(paramTicket, "", paramTicketDefault, paramTicketDesc)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.Remote, false, params.RemoteDesc)
	flagSet.AddFlagSet(params.GetGHActionsFlagset())
	params.AddCirrusFlags(flagSet)
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

			if err := macpool.Release(
				&maptContext.ContextArgs{
					Debug:      viper.IsSet(params.Debug),
					DebugLevel: viper.GetUint(params.DebugLevel),
					Serverless: viper.IsSet(params.Serverless),
					Remote:     viper.IsSet(params.Remote),
				},
				&macpool.MachineRequestArgs{
					VPCID:   viper.GetString(paramVPCID),
					SSHSGID: viper.GetString(paramSSHSGID),
				},
				viper.GetString(paramTicket)); err != nil {
				logging.Error(err)
			}
			return nil
		},
	}
	flagSet := pflag.NewFlagSet(awsParams.MACReleaseCmd, pflag.ExitOnError)
	flagSet.StringP(paramVPCID, "", paramVPCIDDefault, paramVPCIDDesc)
	flagSet.StringP(paramSSHSGID, "", paramSSHSGIDDefault, paramSSHSGIDDesc)
	flagSet.StringP(paramTicket, "", paramTicketDefault, paramTicketDesc)
	flagSet.Bool(params.Serverless, false, params.ServerlessDesc)
	flagSet.Bool(params.Remote, false, params.RemoteDesc)
	c.PersistentFlags().AddFlagSet(flagSet)
	return c
}
