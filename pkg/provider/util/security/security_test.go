package security

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestPasswordArgsRequireEveryCharacterClass(t *testing.T) {
	args := passwordArgs()
	for name, got := range map[string]pulumi.IntPtrInput{
		"MinUpper":   args.MinUpper,
		"MinLower":   args.MinLower,
		"MinNumeric": args.MinNumeric,
		"MinSpecial": args.MinSpecial,
	} {
		if got != pulumi.Int(1) {
			t.Errorf("%s = %v, want 1", name, got)
		}
	}
	if args.Length != pulumi.Int(passwordLength) {
		t.Errorf("Length = %v, want %d", args.Length, passwordLength)
	}
}
