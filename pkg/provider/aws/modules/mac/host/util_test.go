package host

import "testing"

func TestAppendPathToBackedURL(t *testing.T) {
	tests := []struct {
		name, base, sub, want string
	}{
		{"plain s3", "s3://bucket/path", "run1", "s3://bucket/path/run1"},
		{"trailing slash", "s3://bucket/path/", "run1", "s3://bucket/path/run1"},
		{"with query", "s3://bucket/path?endpoint=host&s3ForcePathStyle=true", "run1", "s3://bucket/path/run1?endpoint=host&s3ForcePathStyle=true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendPathToBackedURL(tt.base, tt.sub); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
