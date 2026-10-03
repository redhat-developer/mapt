package security

import (
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/util"
)

const (
	passwordLength          = 16
	passwordOverrideSpecial = "!#%&*()-_=+[]{}.?"
)

func CreatePassword(ctx *pulumi.Context, name string) (*random.RandomPassword, error) {
	return createPassword(ctx, name, nil)
}

func CreatePasswordAlways(ctx *pulumi.Context, name string) (*random.RandomPassword, error) {
	return createPassword(ctx, name,
		[]pulumi.ResourceOption{pulumi.ReplaceOnChanges([]string{"name"}),
			pulumi.Aliases([]pulumi.Alias{{Name: pulumi.String(name)}})})
}

// passwordArgs requires at least one character of each class so the
// generated password always meets the complexity rules of the providers
// (Azure needs 3 of 4 classes and rejects a purely random 16 char value
// that happens to miss them).
func passwordArgs() *random.RandomPasswordArgs {
	return &random.RandomPasswordArgs{
		Length:          pulumi.Int(passwordLength),
		Special:         pulumi.Bool(true),
		OverrideSpecial: pulumi.String(passwordOverrideSpecial),
		MinUpper:        pulumi.Int(1),
		MinLower:        pulumi.Int(1),
		MinNumeric:      pulumi.Int(1),
		MinSpecial:      pulumi.Int(1),
	}
}

func createPassword(ctx *pulumi.Context, name string,
	options []pulumi.ResourceOption) (*random.RandomPassword, error) {
	return random.NewRandomPassword(ctx,
		util.RandomID(name),
		passwordArgs(),
		options...)
}
