package data

import (
	"testing"

	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
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

// TestFilterByFamily_CapAppliedAfterFilter verifies that when ComputeFamilies
// is set, matching types beyond MaxResults are still reachable by filterByFamily
// (i.e. the cap is applied after filtering, not before). This guards the fix for
// the bug where FilterVerbose capped results before filterByFamily ran, silently
// dropping allowlisted families ranked outside the top MaxResults.
func TestFilterByFamily_CapAppliedAfterFilter(t *testing.T) {
	// Build 25 types: 5 non-matching (d3en) followed by 20 m6i types.
	// If cap were applied before filter, the 5 d3en types would consume cap slots
	// and only 15 m6i types would survive. With cap-after-filter all 20 survive.
	var types []string
	for i := 0; i < 5; i++ {
		types = append(types, "d3en.12xlarge")
	}
	for i := 0; i < 20; i++ {
		types = append(types, "m6i.xlarge")
	}

	got := filterByFamily(types, []string{"m6i"})
	if len(got) != 20 {
		t.Errorf("got %d results, want 20 — cap must be applied after filter, not before", len(got))
	}
}

func TestFilterByFamily_CapNotExceeded(t *testing.T) {
	// When filtered results are fewer than MaxResults, all are returned unchanged.
	got := filterByFamily([]string{"m5.xlarge", "m6i.xlarge"}, []string{"m5", "m6i"})
	if len(got) != 2 {
		t.Errorf("got %v, want [m5.xlarge m6i.xlarge]", got)
	}
}

// TestFilters_MaxResultsOmittedWhenFamiliesSet verifies the production cap-order
// fix: filters() must NOT set MaxResults when ComputeFamilies is non-empty.
// If it did, the selector would drop allowlisted families ranked outside the top
// MaxResults before filterByFamily runs, silently producing empty results.
func TestFilters_MaxResultsOmittedWhenFamiliesSet(t *testing.T) {
	args := &computerequest.ComputeRequestArgs{
		CPUs:            4,
		MemoryGib:       16,
		Arch:            computerequest.Amd64,
		ComputeFamilies: []string{"m5", "m6i"},
	}
	f := filters(args)
	if f.MaxResults != nil {
		t.Errorf("MaxResults must be nil when ComputeFamilies is set, got %d — cap must be applied after filterByFamily", *f.MaxResults)
	}
}

// TestFilters_MaxResultsSetWhenNoFamilies verifies that filters() sets MaxResults
// normally when ComputeFamilies is empty (default path, no family filtering).
func TestFilters_MaxResultsSetWhenNoFamilies(t *testing.T) {
	args := &computerequest.ComputeRequestArgs{
		CPUs:      4,
		MemoryGib: 16,
		Arch:      computerequest.Amd64,
	}
	f := filters(args)
	if f.MaxResults == nil {
		t.Error("MaxResults must be set when ComputeFamilies is empty")
	}
	if *f.MaxResults != computerequest.MaxResults {
		t.Errorf("MaxResults = %d, want %d", *f.MaxResults, computerequest.MaxResults)
	}
}
