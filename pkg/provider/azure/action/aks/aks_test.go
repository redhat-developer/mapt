package aks

import (
	"testing"

	containerservice "github.com/pulumi/pulumi-azure-native-sdk/containerservice/v3"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

func stringPtr(s string) *string { return &s }
func intPtr(i int) *int          { return &i }
func boolPtr(b bool) *bool       { return &b }

func newTestRequest() *aksRequest {
	return &aksRequest{
		mCtx:             mc.InitNoState(),
		location:         stringPtr("West US"),
		vmSizes:          []string{"Standard_D4as_v5"},
		systemPoolVMSize: stringPtr(DefaultSystemPoolVMSize),
		workersDesired:   intPtr(DefaultWorkersDesired),
		onlySystemPool:   boolPtr(false),
	}
}

func TestUserPoolSelectionWithoutSpot(t *testing.T) {
	r := newTestRequest()
	selection, err := r.userPoolSelection()
	if err != nil {
		t.Fatalf("userPoolSelection() returned unexpected error: %v", err)
	}
	if selection.location != "West US" {
		t.Errorf("location = %q, want %q", selection.location, "West US")
	}
	if selection.vmSize != "Standard_D4as_v5" {
		t.Errorf("vmSize = %q, want %q", selection.vmSize, "Standard_D4as_v5")
	}
	if selection.price != nil {
		t.Errorf("price = %v, want nil for on demand pools", *selection.price)
	}
}

func TestUserPoolSelectionOnlySystemPool(t *testing.T) {
	r := newTestRequest()
	r.onlySystemPool = boolPtr(true)
	selection, err := r.userPoolSelection()
	if err != nil {
		t.Fatalf("userPoolSelection() returned unexpected error: %v", err)
	}
	if selection.location != "West US" {
		t.Errorf("location = %q, want %q", selection.location, "West US")
	}
	if selection.vmSize != "" {
		t.Errorf("vmSize = %q, want empty as no user pool is created", selection.vmSize)
	}
	if selection.price != nil {
		t.Errorf("price = %v, want nil as no spot search runs without a user pool", *selection.price)
	}
}

func TestUserPoolSelectionFirstVMSizeWithoutSpot(t *testing.T) {
	r := newTestRequest()
	r.vmSizes = []string{"Standard_D4as_v5", "Standard_D4s_v5"}
	selection, err := r.userPoolSelection()
	if err != nil {
		t.Fatalf("userPoolSelection() returned unexpected error: %v", err)
	}
	if selection.vmSize != "Standard_D4as_v5" {
		t.Errorf("vmSize = %q, want the first configured size %q", selection.vmSize, "Standard_D4as_v5")
	}
}

func TestAgentPoolsRegularUserPool(t *testing.T) {
	r := newTestRequest()
	pools := r.agentPools(&poolSelection{
		location: "West US",
		vmSize:   "Standard_D4as_v5",
	})
	if len(pools) != 2 {
		t.Fatalf("len(pools) = %d, want 2", len(pools))
	}
	system, user := pools[0], pools[1]
	if system.name != "systempool" || system.mode != string(containerservice.AgentPoolModeSystem) {
		t.Errorf("system pool = %+v, want name systempool and system mode", system)
	}
	if system.vmSize != DefaultSystemPoolVMSize {
		t.Errorf("system pool vmSize = %q, want %q", system.vmSize, DefaultSystemPoolVMSize)
	}
	if system.count != 1 {
		t.Errorf("system pool count = %d, want 1", system.count)
	}
	if user.name != "userpool" || user.mode != string(containerservice.AgentPoolModeUser) {
		t.Errorf("user pool = %+v, want name userpool and user mode", user)
	}
	if user.vmSize != "Standard_D4as_v5" {
		t.Errorf("user pool vmSize = %q, want %q", user.vmSize, "Standard_D4as_v5")
	}
	if user.spot {
		t.Error("user pool spot = true, want false when no spot price is selected")
	}
	if user.spotMaxPrice != nil {
		t.Errorf("user pool spotMaxPrice = %v, want nil", *user.spotMaxPrice)
	}
}

func TestAgentPoolsSpotUserPool(t *testing.T) {
	r := newTestRequest()
	price := 0.12
	pools := r.agentPools(&poolSelection{
		location: "eastus",
		vmSize:   "Standard_D4as_v5",
		price:    &price,
	})
	if len(pools) != 2 {
		t.Fatalf("len(pools) = %d, want 2", len(pools))
	}
	user := pools[1]
	if !user.spot {
		t.Error("user pool spot = false, want true when a spot price is selected")
	}
	if user.spotMaxPrice == nil || *user.spotMaxPrice != price {
		t.Errorf("user pool spotMaxPrice = %v, want %v", user.spotMaxPrice, price)
	}
	if pools[0].spot {
		t.Error("system pool spot = true, want false as the system pool always runs on demand")
	}
}

func TestAgentPoolsOnlySystemPool(t *testing.T) {
	r := newTestRequest()
	r.onlySystemPool = boolPtr(true)
	pools := r.agentPools(&poolSelection{vmSize: "Standard_D4as_v5"})
	if len(pools) != 1 {
		t.Fatalf("len(pools) = %d, want 1", len(pools))
	}
	if pools[0].name != "systempool" {
		t.Errorf("pool name = %q, want %q", pools[0].name, "systempool")
	}
}

func TestAgentPoolsSystemPoolVMSize(t *testing.T) {
	r := newTestRequest()
	r.systemPoolVMSize = stringPtr("Standard_D2as_v5")
	pools := r.agentPools(&poolSelection{vmSize: "Standard_D4as_v5"})
	if pools[0].vmSize != "Standard_D2as_v5" {
		t.Errorf("system pool vmSize = %q, want %q", pools[0].vmSize, "Standard_D2as_v5")
	}
}

func TestAgentPoolsWorkersDesired(t *testing.T) {
	r := newTestRequest()
	r.workersDesired = intPtr(3)
	pools := r.agentPools(&poolSelection{vmSize: "Standard_D4as_v5"})
	if pools[1].count != 3 {
		t.Errorf("user pool count = %d, want 3", pools[1].count)
	}
}

func TestValidateRequiresUserPoolVMSize(t *testing.T) {
	r := newTestRequest()
	r.vmSizes = nil
	if err := r.validate(); err == nil {
		t.Error("validate() returned nil error, want an error when the user pool has no vmsize")
	}
	r.onlySystemPool = boolPtr(true)
	if err := r.validate(); err != nil {
		t.Errorf("validate() returned unexpected error for a system pool only cluster: %v", err)
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name           string
		onlySystemPool bool
		systemPoolSize string
		workersDesired int
		wantErr        bool
	}{
		{"defaults", false, DefaultSystemPoolVMSize, 1, false},
		{"system pool only", true, DefaultSystemPoolVMSize, 0, false},
		{"several workers", false, DefaultSystemPoolVMSize, 3, false},
		{"no workers", false, DefaultSystemPoolVMSize, 0, true},
		{"negative workers", false, DefaultSystemPoolVMSize, -1, true},
		{"empty system pool size", false, "", 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRequest()
			r.onlySystemPool = boolPtr(tc.onlySystemPool)
			r.systemPoolVMSize = stringPtr(tc.systemPoolSize)
			r.workersDesired = intPtr(tc.workersDesired)
			err := r.validate()
			if tc.wantErr && err == nil {
				t.Error("validate() returned nil error, want an error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validate() returned unexpected error: %v", err)
			}
		})
	}
}
