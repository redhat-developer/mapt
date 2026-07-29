package data

import (
	"testing"
)

func TestFilterByFamily_EmptyFamilies_ReturnsAll(t *testing.T) {
	types := []string{"m5.xlarge", "d3en.12xlarge", "p3.2xlarge"}
	got := filterByFamily(types, nil)
	if len(got) != 3 {
		t.Errorf("got %v, want all 3 types", got)
	}
}

func TestFilterByFamily_SingleFamily(t *testing.T) {
	types := []string{"m5.xlarge", "m5.2xlarge", "d3en.12xlarge", "p3.2xlarge"}
	got := filterByFamily(types, []string{"m5"})
	if len(got) != 2 {
		t.Fatalf("got %v, want [m5.xlarge m5.2xlarge]", got)
	}
	if got[0] != "m5.xlarge" || got[1] != "m5.2xlarge" {
		t.Errorf("got %v, want [m5.xlarge m5.2xlarge]", got)
	}
}

func TestFilterByFamily_MultipleFamilies(t *testing.T) {
	types := []string{"m5.xlarge", "m6i.2xlarge", "d3en.12xlarge", "r5.large"}
	got := filterByFamily(types, []string{"m5", "r5"})
	if len(got) != 2 {
		t.Fatalf("got %v, want [m5.xlarge r5.large]", got)
	}
}

func TestFilterByFamily_NoPrefixSubstringMatch(t *testing.T) {
	// "m5" must NOT match "m5a.8xlarge" — requires dot separator
	types := []string{"m5.xlarge", "m5a.8xlarge"}
	got := filterByFamily(types, []string{"m5"})
	if len(got) != 1 || got[0] != "m5.xlarge" {
		t.Errorf("got %v, want only [m5.xlarge] — m5a must not match m5", got)
	}
}

func TestFilterByFamily_AllFiltered_ReturnsNil(t *testing.T) {
	types := []string{"d3en.12xlarge", "p3.2xlarge"}
	got := filterByFamily(types, []string{"m5"})
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestFilterByFamily_EmptyInput_ReturnsNil(t *testing.T) {
	got := filterByFamily(nil, []string{"m5"})
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}
