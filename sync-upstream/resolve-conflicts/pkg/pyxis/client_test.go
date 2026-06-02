package pyxis

import "testing"

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		ref      string
		wantRepo string
	}{
		{"registry.redhat.io/ubi9/go-toolset:latest", "ubi9/go-toolset"},
		{"registry.redhat.io/ubi9/go-toolset", "ubi9/go-toolset"},
		{"docker://registry.redhat.io/ubi9/go-toolset:latest", "ubi9/go-toolset"},
		{"registry.redhat.io/ubi9/go-toolset:1.21", "ubi9/go-toolset"},
	}
	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			_, repo, err := parseImageRef(tt.ref)
			if err != nil {
				t.Fatalf("parseImageRef: %v", err)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo: got %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}
