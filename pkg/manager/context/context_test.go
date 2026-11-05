package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeProvider struct{ calls int }

func (p *fakeProvider) Init(context.Context, string) (string, error) {
	p.calls++
	return "", nil
}

func (p *fakeProvider) DefaultHostingPlace() (*string, error) {
	p.calls++
	return nil, nil
}

// InitNoState is used by callers that only query cloud APIs, such as
// spot.GetLowestPrice. Those callers pass Context() straight to the cloud
// SDKs, which panic or fail on a nil context.Context.
func TestInitNoStateProvidesContext(t *testing.T) {
	c := InitNoState()
	if c.Context() == nil {
		t.Fatal("InitNoState() returned a Context with a nil context.Context")
	}
	if err := c.Context().Err(); err != nil {
		t.Fatalf("expected a live context, got error: %v", err)
	}
}

// Results are written after the cloud resources exist, so an unusable folder
// has to stop the run before the provider is touched.
func TestInitRejectsUnusableResultsOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	readOnly := filepath.Join(dir, "read-only")
	if err := os.Mkdir(readOnly, 0500); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"missing":      filepath.Join(dir, "missing"),
		"not a folder": file,
	}
	if os.Geteuid() != 0 {
		cases["read only"] = readOnly
	}
	for name, folder := range cases {
		t.Run(name, func(t *testing.T) {
			provider := &fakeProvider{}
			_, err := Init(&ContextArgs{Context: context.Background(), ResultsOutput: folder}, provider)
			if err == nil {
				t.Fatalf("expected an error for %s", folder)
			}
			if provider.calls != 0 {
				t.Errorf("provider was called %d times before the check failed", provider.calls)
			}
		})
	}
}

func TestInitAcceptsUsableResultsOutput(t *testing.T) {
	dir := t.TempDir()
	if _, err := Init(&ContextArgs{Context: context.Background(), ResultsOutput: dir}, &fakeProvider{}); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("the check left files behind: %v", entries)
	}
}

func TestInitSkipsCheckWithoutResultsOutput(t *testing.T) {
	// An empty folder would otherwise fall back to the system temp folder.
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
	if _, err := Init(&ContextArgs{Context: context.Background()}, &fakeProvider{}); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
}
