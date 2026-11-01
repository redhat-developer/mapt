package context

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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
