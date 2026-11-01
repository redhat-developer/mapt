package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
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

func TestResourceTagsWithCustomPerResource(t *testing.T) {
	c := &Context{tags: map[string]string{"projectName": "demo", "owner": "base"}}
	first := c.ResourceTagsWithCustom(map[string]string{"resource": "first", "owner": "override"})
	second := c.ResourceTagsWithCustom(map[string]string{"resource": "second"})
	base := c.ResourceTags()
	if first["resource"] != pulumi.String("first") {
		t.Errorf("first tags: %v", first)
	}
	if second["resource"] != pulumi.String("second") {
		t.Errorf("second tags: %v", second)
	}
	if second["owner"] != pulumi.String("base") {
		t.Errorf("override leaked: %v", second)
	}
	if _, ok := base["resource"]; ok {
		t.Errorf("custom tag leaked to base: %v", base)
	}
	first["projectName"] = pulumi.String("mutated")
	if c.ResourceTags()["projectName"] != pulumi.String("demo") {
		t.Error("returned map mutation leaked")
	}
}

func TestResourceTagsWithCustomCallOrder(t *testing.T) {
	for _, baseFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "custom-first", true: "base-first"}[baseFirst], func(t *testing.T) {
			c := &Context{tags: map[string]string{"owner": "base"}}
			if baseFirst {
				c.ResourceTags()
			}
			custom := map[string]string{"owner": "override", "resource": "specific"}
			got := c.ResourceTagsWithCustom(custom)
			custom["resource"] = "changed-input"
			if got["resource"] != pulumi.String("specific") || got["owner"] != pulumi.String("override") {
				t.Errorf("custom tags lost: %v", got)
			}
			if _, ok := c.ResourceTags()["resource"]; ok {
				t.Error("custom tag leaked to base")
			}
			if c.tags["owner"] != "base" {
				t.Error("base input mutated")
			}
		})
	}
}

func TestResourceTagsWithCustomNilBase(t *testing.T) {
	c := &Context{}
	if len(c.ResourceTags()) != 0 {
		t.Error("nil base should give empty tags")
	}
	if got := c.ResourceTagsWithCustom(map[string]string{"resource": "one"}); got["resource"] != pulumi.String("one") {
		t.Errorf("custom with nil base: %v", got)
	}
}
