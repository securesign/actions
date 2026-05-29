package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/conflict"
)

func TestIsActionBump(t *testing.T) {
	tests := []struct {
		name  string
		block conflict.Block
		want  bool
	}{
		{
			"action bump",
			conflict.Block{
				Ours:   "      - uses: actions/setup-go@aaa # v6.4.0",
				Theirs: "      - uses: actions/setup-go@bbb # v6.3.0",
			},
			true,
		},
		{
			"non-action line",
			conflict.Block{
				Ours:   "      run: echo hello",
				Theirs: "      run: echo world",
			},
			false,
		},
		{
			"mixed",
			conflict.Block{
				Ours:   "      - uses: actions/setup-go@aaa # v6.4.0\n      run: echo hi",
				Theirs: "      - uses: actions/setup-go@bbb # v6.3.0",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsActionBump(tt.block); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveActionBump(t *testing.T) {
	block := conflict.Block{
		Ours:   "      - uses: actions/setup-go@aaa111 # v6.4.0",
		Theirs: "      - uses: actions/setup-go@bbb222 # v6.3.0",
	}

	got := ResolveActionBump(block)
	if !strings.Contains(got, "v6.4.0") {
		t.Errorf("expected v6.4.0 (newer), got %q", got)
	}
}

func TestResolveActionBumpTheirsNewer(t *testing.T) {
	block := conflict.Block{
		Ours:   "      - uses: codecov/codecov-action@aaa # v5.5.2",
		Theirs: "      - uses: codecov/codecov-action@bbb # v6.0.0",
	}

	got := ResolveActionBump(block)
	if !strings.Contains(got, "v6.0.0") {
		t.Errorf("expected v6.0.0 (newer), got %q", got)
	}
}

func TestResolveFile(t *testing.T) {
	input := `name: CI
on: push
jobs:
  test:
    steps:
<<<<<<< HEAD
      - uses: actions/setup-go@aaa111 # v6.4.0
=======
      - uses: actions/setup-go@bbb222 # v6.3.0
>>>>>>> origin/main
      - run: go test ./...
`
	dir := t.TempDir()
	file := filepath.Join(dir, "workflow.yml")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("conflict markers remain")
	}
	if !strings.Contains(string(got), "v6.4.0") {
		t.Error("expected v6.4.0 in output")
	}
	if !strings.Contains(string(got), "go test") {
		t.Error("non-conflict content missing")
	}
}

func TestResolveFileNonActionConflict(t *testing.T) {
	input := `<<<<<<< HEAD
      run: echo hello
=======
      run: echo world
>>>>>>> origin/main
`
	dir := t.TempDir()
	file := filepath.Join(dir, "workflow.yml")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err == nil {
		t.Error("expected error for non-action conflict")
	}
}
