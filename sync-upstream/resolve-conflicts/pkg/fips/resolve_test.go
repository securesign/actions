package fips

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/conflict"
)

func TestIsFIPSConflict(t *testing.T) {
	tests := []struct {
		name string
		block conflict.Block
		want  bool
	}{
		{
			name: "full banner in theirs",
			block: conflict.Block{
				Ours: "upstream code",
				Theirs: `// RHTAS FIPS - DO NOT REMOVE
// ========================================
crypto/fips140.Enabled()
// ========================================`,
			},
			want: true,
		},
		{
			name: "bare comment in theirs",
			block: conflict.Block{
				Ours:   "upstream code",
				Theirs: "// RHTAS FIPS - DO NOT REMOVE\nsome fips code",
			},
			want: true,
		},
		{
			name: "marker in ours",
			block: conflict.Block{
				Ours:   "// RHTAS FIPS - DO NOT REMOVE\nsome code",
				Theirs: "other code",
			},
			want: true,
		},
		{
			name: "no marker",
			block: conflict.Block{
				Ours:   "upstream code",
				Theirs: "downstream code",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFIPSConflict(tt.block); got != tt.want {
				t.Errorf("IsFIPSConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolve_AllFIPS(t *testing.T) {
	input := `package main

<<<<<<< HEAD
func doStuff() {
	hash := sha256.Sum256(data)
}
=======
// RHTAS FIPS - DO NOT REMOVE
// ========================================
func doStuff() {
	if fips140.Enabled() {
		hash := sha256.Sum256(data)
	}
}
// ========================================
>>>>>>> origin/downstream
`

	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	os.WriteFile(file, []byte(input), 0644)

	hasFIPS, fullyResolved, err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !hasFIPS {
		t.Error("expected hasFIPS=true")
	}
	if !fullyResolved {
		t.Error("expected fullyResolved=true")
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("output still contains conflict markers")
	}
	if !strings.Contains(string(got), "RHTAS FIPS - DO NOT REMOVE") {
		t.Error("output missing FIPS marker")
	}
	if !strings.Contains(string(got), "fips140.Enabled()") {
		t.Error("output missing FIPS code")
	}
}

func TestResolve_MixedConflicts(t *testing.T) {
	input := `package main

<<<<<<< HEAD
func doStuff() {
	hash := sha256.Sum256(data)
}
=======
// RHTAS FIPS - DO NOT REMOVE
func doStuff() {
	if fips140.Enabled() {
		hash := sha256.Sum256(data)
	}
}
>>>>>>> origin/downstream
middle section
<<<<<<< HEAD
FROM golang:1.22
=======
FROM golang:1.21
>>>>>>> origin/downstream
end`

	dir := t.TempDir()
	file := filepath.Join(dir, "mixed.go")
	os.WriteFile(file, []byte(input), 0644)

	hasFIPS, fullyResolved, err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !hasFIPS {
		t.Error("expected hasFIPS=true")
	}
	if fullyResolved {
		t.Error("expected fullyResolved=false for mixed conflicts")
	}

	got, _ := os.ReadFile(file)
	if !strings.Contains(string(got), "RHTAS FIPS - DO NOT REMOVE") {
		t.Error("FIPS block should be resolved (downstream taken)")
	}
	if !strings.Contains(string(got), "<<<<<<< HEAD") {
		t.Error("non-FIPS conflict should still have markers")
	}
	if !strings.Contains(string(got), "FROM golang:1.22") {
		t.Error("non-FIPS ours side should be preserved in markers")
	}
	if !strings.Contains(string(got), "FROM golang:1.21") {
		t.Error("non-FIPS theirs side should be preserved in markers")
	}
}

func TestResolve_NoFIPS(t *testing.T) {
	input := `before
<<<<<<< HEAD
upstream change
=======
downstream change
>>>>>>> origin/downstream
after`

	dir := t.TempDir()
	file := filepath.Join(dir, "nofips.txt")
	os.WriteFile(file, []byte(input), 0644)

	hasFIPS, _, err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if hasFIPS {
		t.Error("expected hasFIPS=false")
	}
}

func TestResolve_NoConflicts(t *testing.T) {
	input := "clean file\nno conflicts\n"

	dir := t.TempDir()
	file := filepath.Join(dir, "clean.go")
	os.WriteFile(file, []byte(input), 0644)

	hasFIPS, _, err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if hasFIPS {
		t.Error("expected hasFIPS=false for clean file")
	}
}

func TestResolve_BareComment(t *testing.T) {
	input := `before
<<<<<<< HEAD
original line
=======
original line
// RHTAS FIPS - DO NOT REMOVE
>>>>>>> origin/downstream
after`

	dir := t.TempDir()
	file := filepath.Join(dir, "bare.go")
	os.WriteFile(file, []byte(input), 0644)

	hasFIPS, fullyResolved, err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !hasFIPS {
		t.Error("expected hasFIPS=true")
	}
	if !fullyResolved {
		t.Error("expected fullyResolved=true")
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("output still contains conflict markers")
	}
	if !strings.Contains(string(got), "RHTAS FIPS - DO NOT REMOVE") {
		t.Error("output missing bare FIPS comment")
	}
}
