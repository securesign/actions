package dockerfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/conflict"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/gitutil"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/semver"
)

var imageVersionRe = regexp.MustCompile(`(?:FROM|image:)\s+\S+:v?([0-9]+\.[0-9]+(?:\.[0-9]+)*)(?:[@\s\-]|$)`)

func IsMatch(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.HasPrefix(base, "dockerfile") || strings.HasPrefix(base, "docker-compose")
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
			fmt.Printf("  SKIP %s (dockerfile: %v)\n", f, err)
			failed = append(failed, f)
		} else {
			fmt.Printf("  OK   %s (dockerfile)\n", f)
			if err := gitutil.Add(f); err != nil {
				fmt.Printf("  WARN %s: %v\n", f, err)
			}
			resolved = append(resolved, f)
		}
	}
	return
}

func ExtractImageVersion(content string) string {
	m := imageVersionRe.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	return m[1]
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
		if !isImageVersionConflict(block) {
			return fmt.Errorf("non-image-version conflict found, cannot auto-resolve")
		}
	}

	resolved := f.Render(resolveImageConflict)
	return os.WriteFile(inputPath, []byte(resolved), 0644)
}

func isImageVersionConflict(block conflict.Block) bool {
	for _, side := range []string{block.Ours, block.Theirs} {
		for _, line := range strings.Split(side, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if !isImageLine(line) {
				return false
			}
		}
	}
	return true
}

func isImageLine(line string) bool {
	return strings.HasPrefix(line, "FROM ") || strings.HasPrefix(line, "image:")
}

func resolveImageConflict(block conflict.Block) string {
	oursVer := ExtractImageVersion(block.Ours)
	theirsVer := ExtractImageVersion(block.Theirs)

	if oursVer == "" || theirsVer == "" {
		return block.Ours
	}

	if semver.Compare(oursVer, theirsVer) >= 0 {
		return block.Ours
	}
	return block.Theirs
}
