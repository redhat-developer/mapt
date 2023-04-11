package plugin

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

type SetCredentials func(ctx context.Context, stack auto.Stack, fixedCredentials map[string]string) error

type PluginInfo struct {
	Name              string
	Version           string
	SetCredentialFunc SetCredentials
	FixedCredentials  map[string]string
}

// Each functional plugin should be defined here
// matching versions used from go.mod file (client side of the plugin)
var plugins = []PluginInfo{
	{Name: "command", Version: "v0.7.1"},
	{Name: "random", Version: "v4.11.2"},
	{Name: "tls", Version: " v4.10.0"},
}

func InstallFuntionalPlugins(ctx context.Context, stack *auto.Stack) (err error) {
	for _, p := range plugins {
		if err = stack.Workspace().InstallPlugin(ctx, p.Name, p.Version); err != nil {
			return
		}
	}
	return
}

func InstallCloudProviderPlugin(ctx context.Context, stack *auto.Stack, p PluginInfo) (err error) {
	w := stack.Workspace()
	// for inline source programs, we must manage plugins ourselves
	if err = w.InstallPlugin(ctx, p.Name, p.Version); err != nil {
		return
	}
	// Set credentials
	if p.SetCredentialFunc != nil {
		err = p.SetCredentialFunc(ctx, *stack, p.FixedCredentials)
	}
	return
}
