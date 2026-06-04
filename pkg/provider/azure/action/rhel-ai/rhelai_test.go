package rhelai

import "testing"

func TestIsGPUCapableSize(t *testing.T) {
	cases := []struct {
		size     string
		expected bool
	}{
		{"Standard_ND96asr_v4", true},
		{"Standard_ND40rs_v2", true},
		{"Standard_NC6s_v3", true},
		{"Standard_NC24rs_v3", true},
		{"standard_nd96asr_v4", true},
		{"standard_nc6s_v3", true},
		{"Standard_D8as_v5", false},
		{"Standard_E16as_v5", false},
		{"Standard_F32s_v2", false},
		{"Standard_NV6", false},
		{"Standard_NV36ads_A10_v5", false},
		{"", false},
	}
	for _, tc := range cases {
		got := isGPUCapableSize(tc.size)
		if got != tc.expected {
			t.Errorf("isGPUCapableSize(%q) = %v, want %v", tc.size, got, tc.expected)
		}
	}
}
