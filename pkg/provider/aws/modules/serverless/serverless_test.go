package serverless

import (
	"testing"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

func TestNewOneTimeRequest_passesValidation(t *testing.T) {
	r := newOneTimeRequest(&mc.Context{}, "us-east-1", "main", "snc",
		"aws openshift-snc destroy", "2026-01-15T10:00:00")
	if err := r.validate(); err != nil {
		t.Fatalf("a one-time request must pass validation: %v", err)
	}
}
