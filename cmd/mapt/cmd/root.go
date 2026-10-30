package cmd

import (
	"context"
	"fmt"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/aws"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/azure"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/gcp"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/ibmcloud"
	"github.com/redhat-developer/mapt/cmd/mapt/cmd/params"
	"github.com/redhat-developer/mapt/pkg/telemetry"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	commandName      = "mapt"
	descriptionShort = "Multi Architecture Provisioning Tool"
	descriptionLong  = "MAPT is a tool for creating pre-configured machines (baremetal or VMs) on cloud providers"

	defaultErrorExitCode = 1
)

var (
	baseDir           = filepath.Join(os.Getenv("HOME"), ".mapt")
	logFile           = "mapt.log"
	noTelemetryFlag   bool
	telemetryClient   = telemetry.New(filepath.Join(baseDir, "telemetry.json"))
	telemetryDecision telemetry.Decision
	telemetryDone     <-chan struct{}
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
	decision, err := telemetryClient.Resolve(telemetry.ReadOverrides(noTelemetryFlag), cmd.CommandPath(), telemetryCounted(cmd))
	if err != nil {
		logging.Debugf("telemetry decision unavailable: %v", err)
		telemetryDecision = telemetry.Decision{}
		return nil
	}
	telemetryDecision = decision
	return nil
}

// telemetryCounted reports whether running this command counts as usage.
// attachMiddleware only wraps leaf commands that have a RunE, and generated or
// telemetry-management commands are excluded, so the consent prompt must match
// exactly that set: prompting for a command that is never counted would ask
// for consent and give nothing in return.
func telemetryCounted(cmd *cobra.Command) bool {
	return !cmd.HasSubCommands() && cmd.RunE != nil && !isExcludedTelemetryCommand(cmd.CommandPath())
}

func runRoot() {
	fmt.Println("No command given")
}

func init() {
	// Common flags
	flagSet := pflag.NewFlagSet(commandName, pflag.ExitOnError)
	params.AddDebugFlags(flagSet)
	rootCmd.PersistentFlags().AddFlagSet(flagSet)
	rootCmd.PersistentFlags().BoolVar(&noTelemetryFlag, "no-telemetry", false, "Disable anonymous usage statistics for this run")
	// Subcommands
	rootCmd.AddCommand(
		aws.GetCmd(),
		azure.GetCmd(),
		gcp.GetCmd(),
		ibmcloud.GetCmd())
}

func Execute() {
	attachMiddleware([]string{}, rootCmd)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
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
		err := input(cmd, args)
		if telemetryDecision.Enabled && !isExcludedTelemetryCommand(fullCmd) {
			telemetryDone = telemetryClient.Emit(telemetry.Event{
				InstallID: telemetryDecision.InstallID,
				Version:   telemetry.Version(),
				Command:   fullCmd,
			})
		}
		return err
	}
}

func runPostrun() {
	telemetryClient.Wait(telemetryDone, 2*time.Second)
	telemetryDone = nil
	logging.CloseLogging()
}

func isExcludedTelemetryCommand(commandPath string) bool {
	for _, part := range strings.Fields(commandPath) {
		switch part {
		case "help", "completion", "version":
			return true
		}
	}
	return false
}
