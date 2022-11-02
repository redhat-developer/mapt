package util

import (
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const passwordOverrideSpecial = "!#%&*()-_=+[]{}<>.?"

func CreatePassword(ctx *pulumi.Context, name string) (*random.RandomPassword, error) {
	return random.NewRandomPassword(ctx,
		name,
		&random.RandomPasswordArgs{
			Length:          pulumi.Int(16),
			Special:         pulumi.Bool(true),
			OverrideSpecial: pulumi.String(passwordOverrideSpecial),
		})
}
