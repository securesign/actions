package fips

import (
	"fmt"
	"os"
	"strings"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/conflict"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/gitutil"
)

const Marker = "RHTAS FIPS - DO NOT REMOVE"

func IsFIPSConflict(block conflict.Block) bool {
	return strings.Contains(block.Ours, Marker) || strings.Contains(block.Theirs, Marker)
}

func ResolveAll() (resolved, partiallyResolved, failed []string, err error) {
	files, err := gitutil.ConflictingFiles()
	if err != nil {
		return nil, nil, nil, err
	}
	for _, f := range files {
		hasFIPS, fullyResolved, resolveErr := Resolve(f)
		if !hasFIPS {
			continue
		}
		if resolveErr != nil {
			fmt.Printf("  SKIP %s (fips: %v)\n", f, resolveErr)
			failed = append(failed, f)
			continue
		}
		if fullyResolved {
			fmt.Printf("  OK   %s (fips)\n", f)
			if err := gitutil.Add(f); err != nil {
				fmt.Printf("  WARN %s: %v\n", f, err)
			}
			resolved = append(resolved, f)
		} else {
			fmt.Printf("  PARTIAL %s (fips: non-FIPS conflicts remain)\n", f)
			partiallyResolved = append(partiallyResolved, f)
		}
	}
	return
}

// Resolve processes a single file. It returns whether the file contained FIPS
// conflicts and whether all conflicts were resolved.
func Resolve(path string) (hasFIPS, fullyResolved bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false, fmt.Errorf("reading file: %w", err)
	}

	f, err := conflict.Parse(strings.NewReader(string(data)))
	if err != nil {
		return false, false, fmt.Errorf("parsing conflicts: %w", err)
	}

	if !f.HasConflicts() {
		return false, false, nil
	}

	fipsCount := 0
	for _, block := range f.Conflicts() {
		if IsFIPSConflict(block) {
			fipsCount++
		}
	}
	if fipsCount == 0 {
		return false, false, nil
	}

	rendered := f.Render(func(block conflict.Block) string {
		if IsFIPSConflict(block) {
			return block.Theirs
		}
		return block.ReEmit()
	})

	if err := os.WriteFile(path, []byte(rendered), 0644); err != nil {
		return true, false, fmt.Errorf("writing file: %w", err)
	}

	fullyResolved = fipsCount == len(f.Conflicts())
	return true, fullyResolved, nil
}
