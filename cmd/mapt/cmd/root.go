package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/mapt/cmd/mapt/cmd/aws"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/azure"
	params "github.com/redhat-developer/mapt/cmd/mapt/cmd/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	commandName      = "mapt"
	descriptionShort = "Multi Architecture Provisioning Tool"
	descriptionLong  = "MAPT is a tool for creating pre-configured machines (baremetal or VMs) on cloud providers"

	defaultErrorExitCode = 1
)

var (
	baseDir = filepath.Join(os.Getenv("HOME"), ".mapt")
	logFile = "mapt.log"
)

var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: descriptionShort,
	Long:  descriptionLong,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return runPrerun(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
		_ = cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runPrerun(cmd *cobra.Command) error {
	logging.InitLogrus(baseDir, logFile)
	return nil
}

func runRoot() {
	fmt.Println("No command given")
}

func init() {
	// Common flags
	flagSet := pflag.NewFlagSet(commandName, pflag.ExitOnError)
	flagSet.StringP(params.ProjectName, "", "", params.ProjectNameDesc)
	flagSet.StringP(params.BackedURL, "", "", params.BackedURLDesc)
	rootCmd.PersistentFlags().AddFlagSet(flagSet)
	// Subcommands
	rootCmd.AddCommand(
		aws.GetCmd(),
		azure.GetCmd())
}

func Execute() {
	attachMiddleware([]string{}, rootCmd)

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		runPostrun()
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(defaultErrorExitCode)
	}
	runPostrun()
}

func attachMiddleware(names []string, cmd *cobra.Command) {
	if cmd.HasSubCommands() {
		for _, command := range cmd.Commands() {
			attachMiddleware(append(names, cmd.Name()), command)
		}
	} else if cmd.RunE != nil {
		fullCmd := strings.Join(append(names, cmd.Name()), " ")
		src := cmd.RunE
		cmd.RunE = executeWithLogging(fullCmd, src)
	}
}

func executeWithLogging(fullCmd string, input func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logging.Debugf("running '%s'", fullCmd)
		return input(cmd, args)
	}
}

func runPostrun() {
	logging.CloseLogging()
}
