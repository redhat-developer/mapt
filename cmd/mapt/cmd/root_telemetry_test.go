package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTelemetryCounted(t *testing.T) {
	root := &cobra.Command{Use: "mapt"}
	aws := &cobra.Command{Use: "aws", RunE: noopRunE}
	hosts := &cobra.Command{Use: "hosts"}
	create := &cobra.Command{Use: "create", RunE: noopRunE}
	list := &cobra.Command{Use: "list"}

	root.AddCommand(aws)
	aws.AddCommand(hosts)
	hosts.AddCommand(create, list)

	tests := []struct {
		name string
		cmd  *cobra.Command
		want bool
	}{
		{"root shows help", root, false},
		{"group with RunE", aws, false},
		{"intermediate group", hosts, false},
		{"leaf with RunE", create, true},
		{"leaf without RunE shows help", list, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := telemetryCounted(tt.cmd); got != tt.want {
				t.Fatalf("telemetryCounted(%q) = %t, want %t", tt.cmd.CommandPath(), got, tt.want)
			}
		})
	}
}

func noopRunE(cmd *cobra.Command, args []string) error { return nil }
