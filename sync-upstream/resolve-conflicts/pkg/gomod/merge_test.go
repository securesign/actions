package gomod

import (
	"testing"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestMergeGoVersion(t *testing.T) {
	tests := []struct {
		name, ours, theirs, ceiling, want string
	}{
		{"theirs newer within ceiling", "1.25.7", "1.26.2", "1.26.2", "1.26.2"},
		{"theirs newer above ceiling", "1.25.7", "1.26.3", "1.26.2", "1.26.2"},
		{"ours newer", "1.26.1", "1.25.0", "1.26.2", "1.26.1"},
		{"no ceiling", "1.25.7", "1.26.3", "", "1.26.3"},
		{"same version", "1.25.7", "1.25.7", "1.26.2", "1.25.7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ours := goModWithVersion(t, tt.ours)
			theirs := goModWithVersion(t, tt.theirs)
			got := mergeGoVersion(ours, theirs, tt.ceiling)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name, baseVer, oursVer, theirsVer, wantVer string
	}{
		{"both changed, theirs newer", "v0.25.0", "v0.25.5", "v0.26.0", "v0.26.0"},
		{"both changed, ours newer", "v0.25.0", "v0.27.0", "v0.26.0", "v0.27.0"},
		{"only ours changed", "v0.25.0", "v0.26.0", "v0.25.0", "v0.26.0"},
		{"only theirs changed", "v0.25.0", "v0.25.0", "v0.26.0", "v0.26.0"},
		{"same version", "v1.0.0", "v1.0.0", "v1.0.0", "v1.0.0"},
		{"date-based, ours newer", "v0.0.0-20210101000000-abc123", "v0.0.0-20241212093149-d2f9f4", "v0.0.0-20211028175153-1c139d", "v0.0.0-20241212093149-d2f9f4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ours := &modfile.Require{Mod: module.Version{Path: "example.com/dep", Version: tt.oursVer}}
			theirs := &modfile.Require{Mod: module.Version{Path: "example.com/dep", Version: tt.theirsVer}}
			got := ResolveVersion(tt.baseVer, ours, theirs)
			if got == nil || got.Mod.Version != tt.wantVer {
				t.Errorf("got %v, want %q", got, tt.wantVer)
			}
		})
	}
}

func TestResolveVersionNil(t *testing.T) {
	req := &modfile.Require{Mod: module.Version{Path: "example.com/dep", Version: "v1.0.0"}}
	if got := ResolveVersion("", nil, req); got != req {
		t.Error("expected theirs when ours is nil")
	}
	if got := ResolveVersion("", req, nil); got != req {
		t.Error("expected ours when theirs is nil")
	}
	if got := ResolveVersion("", nil, nil); got != nil {
		t.Error("expected nil when both nil")
	}
}

func TestMergeIntegration(t *testing.T) {
	lax := func(_, v string) (string, error) { return v, nil }
	parse := func(s string) *modfile.File {
		f, err := modfile.Parse("test", []byte(s), lax)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		return f
	}

	base := parse("module test\n\ngo 1.25.0\n\nrequire (\n\tgithub.com/foo/bar v1.0.0\n\tgithub.com/shared/dep v0.5.0\n)\n")
	ours := parse("module test\n\ngo 1.25.7\n\nrequire (\n\tgithub.com/foo/bar v1.1.0\n\tgithub.com/shared/dep v0.6.0\n\tgithub.com/downstream/only v1.0.0\n)\n")
	theirs := parse("module test\n\ngo 1.26.3\n\nrequire (\n\tgithub.com/foo/bar v1.0.0\n\tgithub.com/shared/dep v0.7.0\n\tgithub.com/upstream/only v0.1.0\n)\n")

	result, err := Merge(base, ours, theirs, "1.26.2")
	if err != nil {
		t.Fatalf("merge: %v", err)
	}

	if result.Go.Version != "1.26.2" {
		t.Errorf("go version: %q, want 1.26.2", result.Go.Version)
	}

	reqs := requireMap(result)
	check := func(path, want string) {
		t.Helper()
		r, ok := reqs[path]
		if !ok {
			t.Errorf("missing %s", path)
			return
		}
		if r.Mod.Version != want {
			t.Errorf("%s: got %q, want %q", path, r.Mod.Version, want)
		}
	}
	check("github.com/foo/bar", "v1.1.0")
	check("github.com/shared/dep", "v0.7.0")
	check("github.com/downstream/only", "v1.0.0")
	check("github.com/upstream/only", "v0.1.0")
}

func goModWithVersion(t *testing.T, version string) *modfile.File {
	t.Helper()
	f, err := modfile.Parse("go.mod", []byte("module test\n\ngo "+version+"\n"), nil)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return f
}
