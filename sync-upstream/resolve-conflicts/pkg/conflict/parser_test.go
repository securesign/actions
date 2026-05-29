package conflict

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := `line 1
line 2
<<<<<<< HEAD
      - uses: actions/setup-go@aaa111 # v6.4.0
=======
      - uses: actions/setup-go@bbb222 # v6.3.0
>>>>>>> origin/main
line 3`

	f, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(f.Sections) != 3 {
		t.Fatalf("got %d sections, want 3", len(f.Sections))
	}

	if f.Sections[0].Text != "line 1\nline 2" {
		t.Errorf("section 0: %q", f.Sections[0].Text)
	}

	if f.Sections[1].Conflict == nil {
		t.Fatal("section 1 should be a conflict")
	}
	if !strings.Contains(f.Sections[1].Conflict.Ours, "v6.4.0") {
		t.Errorf("ours: %q", f.Sections[1].Conflict.Ours)
	}
	if !strings.Contains(f.Sections[1].Conflict.Theirs, "v6.3.0") {
		t.Errorf("theirs: %q", f.Sections[1].Conflict.Theirs)
	}

	if f.Sections[2].Text != "line 3" {
		t.Errorf("section 2: %q", f.Sections[2].Text)
	}
}

func TestParseMultipleConflicts(t *testing.T) {
	input := `before
<<<<<<< HEAD
ours1
=======
theirs1
>>>>>>> origin/main
middle
<<<<<<< HEAD
ours2
=======
theirs2
>>>>>>> origin/main
after`

	f, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	conflicts := f.Conflicts()
	if len(conflicts) != 2 {
		t.Fatalf("got %d conflicts, want 2", len(conflicts))
	}

	if conflicts[0].Ours != "ours1" {
		t.Errorf("conflict 0 ours: %q", conflicts[0].Ours)
	}
	if conflicts[1].Theirs != "theirs2" {
		t.Errorf("conflict 1 theirs: %q", conflicts[1].Theirs)
	}
}

func TestRender(t *testing.T) {
	input := `before
<<<<<<< HEAD
ours
=======
theirs
>>>>>>> origin/main
after`

	f, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	got := f.Render(func(b Block) string {
		return b.Ours
	})

	want := "before\nours\nafter\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNoConflicts(t *testing.T) {
	input := "line 1\nline 2\nline 3"
	f, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if f.HasConflicts() {
		t.Error("expected no conflicts")
	}
}
