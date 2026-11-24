package manager

import (
	"context"
	"errors"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
)

type fakeDestroyer struct {
	err  error
	opts optdestroy.Options
}

func (f *fakeDestroyer) Destroy(_ context.Context, opts ...optdestroy.Option) (auto.DestroyResult, error) {
	for _, option := range opts {
		option.ApplyOption(&f.opts)
	}
	return auto.DestroyResult{}, f.err
}

func TestDestroyPhasePreservesState(t *testing.T) {
	for _, tt := range []struct {
		name    string
		opts    DestroyOptions
		failure bool
		removed bool
	}{
		{"final", DestroyOptions{}, false, true},
		{"targeted", DestroyOptions{Targets: []string{"cluster"}, TargetDependents: true, ContinueOnError: true}, false, false},
		{"targeted-no-dependents", DestroyOptions{Targets: []string{"kubernetes:**"}}, false, false},
		{"failed", DestroyOptions{}, true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeDestroyer{}
			if tt.failure {
				f.err = errors.New("dependency violation")
			}
			removed := false
			err := destroyPhase(context.Background(), f, func(context.Context) error { removed = true; return nil }, tt.opts)
			if (err != nil) != tt.failure || removed != tt.removed {
				t.Fatalf("err=%v removed=%v", err, removed)
			}
			if len(tt.opts.Targets) > 0 && (f.opts.TargetDependents != tt.opts.TargetDependents || len(f.opts.Target) != 1) {
				t.Fatalf("target options not applied: %+v", f.opts)
			}
			if f.opts.ContinueOnError != tt.opts.ContinueOnError {
				t.Fatal("continue-on-error lost")
			}
		})
	}
}

func TestDestroyPhaseBackgroundReturnsUnwrappedErrors(t *testing.T) {
	cause := errors.New("dependency violation")
	removeErr := errors.New("backend unavailable")
	noRemove := func(context.Context) error { return nil }
	for _, tt := range []struct {
		name    string
		opts    DestroyOptions
		destroy error
		remove  func(context.Context) error
		wantErr error
	}{
		{"destroy", DestroyOptions{Background: true}, cause, noRemove, cause},
		{"remove", DestroyOptions{Background: true}, nil, func(context.Context) error { return removeErr }, removeErr},
		{"wrapped", DestroyOptions{}, cause, noRemove, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := destroyPhase(context.Background(), &fakeDestroyer{err: tt.destroy}, tt.remove, tt.opts)
			if tt.wantErr != nil && err != tt.wantErr {
				t.Fatalf("err=%v, want the unwrapped %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (!errors.Is(err, cause) || err == cause) {
				t.Fatalf("err=%v, want a wrapped cause", err)
			}
		})
	}
}
