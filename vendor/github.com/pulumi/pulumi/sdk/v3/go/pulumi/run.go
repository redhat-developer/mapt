// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pulumi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	multierror "github.com/hashicorp/go-multierror"

	"github.com/pulumi/pulumi/sdk/v3/go/common/constant"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
)

var ErrPlugins = errors.New("pulumi: plugins requested")

// A RunOption is used to control the behavior of Run and RunErr.
type RunOption func(*RunInfo)

// Run executes the body of a Pulumi program, granting it access to a deployment context that it may use
// to register resources and orchestrate deployment activities.  This connects back to the Pulumi engine using gRPC.
// If the program fails, the process will be terminated and the function will not return.
func Run(body RunFunc, opts ...RunOption) {
	logError := func(ctx *Context, programErr error) {
		logErr := ctx.Log.Error(fmt.Sprintf("an unhandled error occurred: program failed: \n%v",
			programErr), nil)
		contract.IgnoreError(logErr)
	}

	err := runErrInner(body, logError, opts...)
	if err == nil {
		return
	}

	if err == ErrPlugins {
		printRequiredPlugins()
	}

	os.Exit(constant.ExitStatusLoggedError)
}

// RunErr executes the body of a Pulumi program, granting it access to a deployment context that it may use
// to register resources and orchestrate deployment activities.  This connects back to the Pulumi engine using gRPC.
func RunErr(body RunFunc, opts ...RunOption) error {
	return runErrInner(body, func(*Context, error) {}, opts...)
}

func runErrInner(body RunFunc, logError func(*Context, error), opts ...RunOption) error {
	// Parse the info out of environment variables.  This is a lame contract with the caller, but helps to keep
	// boilerplate to a minimum in the average Pulumi Go program.
	info := getEnvInfo()
	if info.getPlugins {
		return ErrPlugins
	}

	for _, o := range opts {
		o(&info)
	}

	// Validate some properties.
	if info.Project == "" {
		return errors.New("missing project name")
	} else if info.Stack == "" {
		return errors.New("missing stack name")
	} else if info.MonitorAddr == "" && info.Mocks == nil {
		return errors.New("missing resource monitor RPC address")
	} else if info.EngineAddr == "" && info.Mocks == nil {
		return errors.New("missing engine RPC address")
	}

	// Create a fresh context.
	ctx, err := NewContext(context.TODO(), info)
	if err != nil {
		return err
	}
	defer contract.IgnoreClose(ctx)

	err = RunWithContext(ctx, body)
	// Log the error message
	if err != nil {
		logError(ctx, err)
	} else {
		if _, signalErr := ctx.state.monitor.SignalAndWaitForShutdown(ctx.ctx, &pbempty.Empty{}); signalErr != nil {
			status, ok := status.FromError(err)
			if ok && status.Code() != codes.Unimplemented {
				// If we are running against an older version of the CLI,
				// SignalAndWaitForShutdown might not be implemented. This is
				// mostly fine, but means that delete hooks do not work. Since
				// we check if the CLI supports the `resourceHook` feature when
				// registering hooks, it's fine to ignore the `UNIMPLEMENTED`
				// error here.
				return fmt.Errorf("error waiting for shutdown: %v", signalErr)
			}
		}
	}

	return err
}

// RunWithContext runs the body of a Pulumi program using the given Context for information about the target stack,
// configuration, and engine connection.
func RunWithContext(ctx *Context, body RunFunc) error {
	info := ctx.state.info

	// Create a root stack resource that we'll parent everything to.
	var stack ResourceState
	err := ctx.RegisterResource(
		"pulumi:pulumi:Stack", fmt.Sprintf("%s-%s", info.Project, info.Stack), nil, &stack)
	if err != nil {
		return err
	}
	ctx.state.stack = &stack

	// Execute the body.
	var result error
	if err = body(ctx); err != nil {
		result = multierror.Append(result, err)
	}

	// Register all the outputs to the stack object.
	if err = ctx.RegisterResourceOutputs(ctx.state.stack, Map(ctx.state.exports)); err != nil {
		result = multierror.Append(result, err)
	}

	if err = ctx.wait(); err != nil {
		return err
	}

	// Propagate the error from the body, if any.
	return result
}

// RunFunc executes the body of a Pulumi program.  It may register resources using the deployment context
// supplied as an arguent and any non-nil return value is interpreted as a program error by the Pulumi runtime.
type RunFunc func(ctx *Context) error

// RunInfo contains all the metadata about a run request.
type RunInfo struct {
	Project           string
	RootDirectory     string
	Stack             string
	Config            map[string]string
	ConfigSecretKeys  []string
	ConfigPropertyMap resource.PropertyMap
	Parallel          int32
	DryRun            bool
	MonitorAddr       string
	EngineAddr        string
	Organization      string
	Mocks             MockResourceMonitor

	getPlugins bool
	engineConn *grpc.ClientConn // Pre-existing engine connection. If set this is used over EngineAddr.

	// If non-nil, wraps the resource monitor client used by Context.
	wrapResourceMonitorClient func(pulumirpc.ResourceMonitorClient) pulumirpc.ResourceMonitorClient
}

// getEnvInfo reads various program information from the process environment.
func getEnvInfo() RunInfo {
	// Most of the variables are just strings, and we can read them directly.  A few of them require more parsing.
	parallel, _ := strconv.ParseInt(os.Getenv(EnvParallel), 10, 32)
	dryRun, _ := strconv.ParseBool(os.Getenv(EnvDryRun))
	getPlugins, _ := strconv.ParseBool(os.Getenv(envPlugins))

	var config map[string]string
	if cfg := os.Getenv(EnvConfig); cfg != "" {
		_ = json.Unmarshal([]byte(cfg), &config)
	}

	var configSecretKeys []string
	if keys := os.Getenv(EnvConfigSecretKeys); keys != "" {
		_ = json.Unmarshal([]byte(keys), &configSecretKeys)
	}

	return RunInfo{
		Organization:     os.Getenv(EnvOrganization),
		Project:          os.Getenv(EnvProject),
		RootDirectory:    os.Getenv(EnvPulumiRootDirectory),
		Stack:            os.Getenv(EnvStack),
		Config:           config,
		ConfigSecretKeys: configSecretKeys,
		Parallel:         int32(parallel), //nolint:gosec // guarded by strconv.ParseInt
		DryRun:           dryRun,
		MonitorAddr:      os.Getenv(EnvMonitor),
		EngineAddr:       os.Getenv(EnvEngine),
		getPlugins:       getPlugins,
	}
}

const (
	// EnvOrganization is the envvar used to read the current Pulumi organization name.
	EnvOrganization = "PULUMI_ORGANIZATION"
	// EnvProject is the envvar used to read the current Pulumi project name.
	EnvProject = "PULUMI_PROJECT"
	// EnvPulumiRootDirectory is the envvar used to read the current Pulumi project root, location of Pulumi.yaml.
	EnvPulumiRootDirectory = "PULUMI_ROOT_DIRECTORY"
	// EnvStack is the envvar used to read the current Pulumi stack name.
	EnvStack = "PULUMI_STACK"
	// EnvConfig is the envvar used to read the current Pulumi configuration variables.
	EnvConfig = "PULUMI_CONFIG"
	// EnvConfigSecretKeys is the envvar used to read the current Pulumi configuration keys that are secrets.
	//nolint:gosec
	EnvConfigSecretKeys = "PULUMI_CONFIG_SECRET_KEYS"
	// EnvParallel is the envvar used to read the current Pulumi degree of parallelism.
	EnvParallel = "PULUMI_PARALLEL"
	// EnvDryRun is the envvar used to read the current Pulumi dry-run setting.
	EnvDryRun = "PULUMI_DRY_RUN"
	// EnvMonitor is the envvar used to read the current Pulumi monitor RPC address.
	EnvMonitor = "PULUMI_MONITOR"
	// EnvEngine is the envvar used to read the current Pulumi engine RPC address.
	EnvEngine = "PULUMI_ENGINE"
	// envPlugins is the envvar used to request that the Pulumi program print its set of required plugins and exit.
	envPlugins = "PULUMI_PLUGINS"
)

type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Server  string `json:"server,omitempty"`
}

var packageRegistry = map[PackageInfo]struct{}{}

func RegisterPackage(info PackageInfo) {
	packageRegistry[info] = struct{}{}
}

func printRequiredPlugins() {
	plugins := []PackageInfo{}
	for info := range packageRegistry {
		plugins = append(plugins, info)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	err := enc.Encode(map[string]interface{}{"plugins": plugins})
	contract.IgnoreError(err)
}
