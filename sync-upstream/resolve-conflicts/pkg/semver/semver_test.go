package semver

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.1.0", "v1.0.0", 1},
		{"v1.0.0", "v1.1.0", -1},
		{"1.26.3", "1.26.2", 1},
		{"1.25.7", "1.26.3", -1},
	}
	for _, tt := range tests {
		got := Compare(tt.a, tt.b)
		if (tt.want > 0 && got <= 0) || (tt.want < 0 && got >= 0) || (tt.want == 0 && got != 0) {
			t.Errorf("Compare(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestNewer(t *testing.T) {
	if got := Newer("1.26.3", "1.25.7"); got != "1.26.3" {
		t.Errorf("got %q, want 1.26.3", got)
	}
	if got := Newer("1.25.7", "1.26.3"); got != "1.26.3" {
		t.Errorf("got %q, want 1.26.3", got)
	}
}

func TestClampToMax(t *testing.T) {
	if got := ClampToMax("1.26.3", "1.26.2"); got != "1.26.2" {
		t.Errorf("got %q, want 1.26.2 (clamped)", got)
	}
	if got := ClampToMax("1.25.7", "1.26.2"); got != "1.25.7" {
		t.Errorf("got %q, want 1.25.7 (under ceiling)", got)
	}
	if got := ClampToMax("1.26.3", ""); got != "1.26.3" {
		t.Errorf("got %q, want 1.26.3 (no ceiling)", got)
	}
}
