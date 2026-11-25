package services

import (
	"reflect"
	"testing"
)

func TestCreateAKSParsesVMSizes(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		expected []string
	}{
		{"default", []string{}, []string{paramVMSizeDefault}},
		{"single size", []string{"--vmsize", "Standard_D4as_v5"}, []string{"Standard_D4as_v5"}},
		{
			"comma separated sizes",
			[]string{"--vmsize", "Standard_D4as_v5,Standard_D4s_v5"},
			[]string{"Standard_D4as_v5", "Standard_D4s_v5"},
		},
		{
			"repeated flag",
			[]string{"--vmsize", "Standard_D4as_v5", "--vmsize", "Standard_D4s_v5"},
			[]string{"Standard_D4as_v5", "Standard_D4s_v5"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := getCreateAKS()
			if err := c.ParseFlags(tc.args); err != nil {
				t.Fatalf("ParseFlags(%v) returned unexpected error: %v", tc.args, err)
			}
			got, err := c.Flags().GetStringSlice(paramVMSize)
			if err != nil {
				t.Fatalf("GetStringSlice(%q) returned unexpected error: %v", paramVMSize, err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("vmsizes = %v, want %v", got, tc.expected)
			}
		})
	}
}
