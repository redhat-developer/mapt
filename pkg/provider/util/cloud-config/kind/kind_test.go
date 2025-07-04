package kind

import (
	"testing"
)

func TestParseExtraPortMappings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []PortMapping
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []PortMapping{},
			wantErr:  false,
		},
		{
			name:  "single port mapping",
			input: `[{"containerPort": 8080, "hostPort": 8080, "protocol": "TCP"}]`,
			expected: []PortMapping{
				{ContainerPort: 8080, HostPort: 8080, Protocol: "TCP"},
			},
			wantErr: false,
		},
		{
			name:  "multiple port mappings",
			input: `[{"containerPort": 8080, "hostPort": 8080, "protocol": "TCP"}, {"containerPort": 9090, "hostPort": 9090, "protocol": "UDP"}]`,
			expected: []PortMapping{
				{ContainerPort: 8080, HostPort: 8080, Protocol: "TCP"},
				{ContainerPort: 9090, HostPort: 9090, Protocol: "UDP"},
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			input:    `[{"containerPort": 8080, "hostPort": 8080, "protocol": "TCP"`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExtraPortMappings(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExtraPortMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.expected) {
				t.Errorf("ParseExtraPortMappings() got %d port mappings, expected %d", len(got), len(tt.expected))
				return
			}
			for i, mapping := range got {
				if i < len(tt.expected) {
					if mapping.ContainerPort != tt.expected[i].ContainerPort ||
						mapping.HostPort != tt.expected[i].HostPort ||
						mapping.Protocol != tt.expected[i].Protocol {
						t.Errorf("ParseExtraPortMappings() got %+v, expected %+v", mapping, tt.expected[i])
					}
				}
			}
		})
	}
}
