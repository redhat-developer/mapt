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
