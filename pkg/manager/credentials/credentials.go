package credentials

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

type SetCredentials func(ctx context.Context, mCtx *mc.Context, stack auto.Stack, fixedCredentials map[string]string) error

type ProviderCredentials struct {
	SetCredentialFunc SetCredentials
	FixedCredentials  map[string]string
}

func SetProviderCredentials(ctx context.Context, mCtx *mc.Context, stack *auto.Stack, p ProviderCredentials) (err error) {
	// Set credentials
	if p.SetCredentialFunc != nil {
		err = p.SetCredentialFunc(ctx, mCtx, *stack, p.FixedCredentials)
	}
	return
}
