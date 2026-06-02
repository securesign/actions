package workflow

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/conflict"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/gitutil"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/semver"
)

var (
	usesRe    = regexp.MustCompile(`uses:\s+\S+`)
	versionRe = regexp.MustCompile(`#\s*v([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)
)

func IsMatch(path string) bool {
	return strings.HasPrefix(path, ".github/workflows/") &&
		(strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml"))
}

func ResolveAll() (resolved, failed []string, err error) {
	files, err := gitutil.ConflictingFiles()
	if err != nil {
		return nil, nil, err
	}
	for _, f := range files {
		if !IsMatch(f) {
			continue
		}
		if err := Resolve(f); err != nil {
			fmt.Printf("  SKIP %s (workflow: %v)\n", f, err)
			failed = append(failed, f)
		} else {
			fmt.Printf("  OK   %s (workflow)\n", f)
			if err := gitutil.Add(f); err != nil {
				fmt.Printf("  WARN %s: %v\n", f, err)
			}
			resolved = append(resolved, f)
		}
	}
	return
}

func IsActionBump(block conflict.Block) bool {
	for _, side := range []string{block.Ours, block.Theirs} {
		for _, line := range strings.Split(side, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if !usesRe.MatchString(line) {
				return false
			}
		}
	}
	return true
}

func ResolveActionBump(block conflict.Block) string {
	oursVer := extractVersion(block.Ours)
	theirsVer := extractVersion(block.Theirs)

	if oursVer == "" || theirsVer == "" {
		return block.Ours
	}

	if semver.Compare(oursVer, theirsVer) >= 0 {
		return block.Ours
	}
	return block.Theirs
}

func Resolve(inputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	f, err := conflict.Parse(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("parsing conflicts: %w", err)
	}

	if !f.HasConflicts() {
		return nil
	}

	for _, block := range f.Conflicts() {
		if !IsActionBump(block) {
			return fmt.Errorf("non-action-bump conflict found, cannot auto-resolve")
		}
		if extractVersion(block.Ours) == "" || extractVersion(block.Theirs) == "" {
			return fmt.Errorf("cannot extract action version from conflict block, resolve manually")
		}
	}

	resolved := f.Render(ResolveActionBump)
	return os.WriteFile(inputPath, []byte(resolved), 0644)
}

func extractVersion(s string) string {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return m[1]
}
