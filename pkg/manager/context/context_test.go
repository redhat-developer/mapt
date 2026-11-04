package context

import "testing"

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
