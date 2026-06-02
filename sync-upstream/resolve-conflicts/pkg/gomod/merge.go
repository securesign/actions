package gomod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/gitutil"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/semver"
	"golang.org/x/mod/modfile"
)

func IsMatch(path string) bool {
	return strings.HasSuffix(path, "go.mod")
}

func ResolveAll(goCeiling string) (resolved, failed []string, err error) {
	files, err := gitutil.ConflictingFiles()
	if err != nil {
		return nil, nil, err
	}
	for _, f := range files {
		if !IsMatch(f) {
			continue
		}
		if err := Resolve(f, goCeiling); err != nil {
			fmt.Printf("  SKIP %s (gomod: %v)\n", f, err)
			failed = append(failed, f)
			continue
		}
		fmt.Printf("  OK   %s (gomod)\n", f)
		if err := gitutil.Add(f); err != nil {
			fmt.Printf("  WARN %s: %v\n", f, err)
		}
		resolved = append(resolved, f)

		dir := filepath.Dir(f)
		sumFile := filepath.Join(dir, "go.sum")
		if data, err := os.ReadFile(sumFile); err == nil && strings.Contains(string(data), "<<<<<<<") {
			os.Remove(sumFile)
			fmt.Printf("  Removed conflicted %s\n", sumFile)
		}

		fmt.Printf("  Running go mod tidy in %s\n", dir)
		tidy := exec.Command("go", "mod", "tidy")
		tidy.Dir = dir
		tidy.Stdout = os.Stdout
		tidy.Stderr = os.Stderr
		if err := tidy.Run(); err != nil {
			fmt.Printf("  WARN go mod tidy failed in %s: %v\n", dir, err)
		}
		gitutil.Add(filepath.Join(dir, "go.mod"))
		gitutil.Add(filepath.Join(dir, "go.sum"))
	}
	return
}

func Resolve(path, goCeiling string) error {
	mergeBase := gitutil.MergeBase()

	tmpdir, err := os.MkdirTemp("", "gomod-resolve-")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	baseFile := filepath.Join(tmpdir, "base.mod")
	oursFile := filepath.Join(tmpdir, "ours.mod")
	theirsFile := filepath.Join(tmpdir, "theirs.mod")

	if err := gitutil.ShowToFile(mergeBase+":"+path, baseFile); err != nil {
		os.WriteFile(baseFile, []byte("module unknown\n"), 0644)
	}
	if err := gitutil.ShowToFile("MERGE_HEAD:"+path, oursFile); err != nil {
		return fmt.Errorf("cannot read MERGE_HEAD:%s — modify/delete conflict", path)
	}
	if err := gitutil.ShowToFile("HEAD:"+path, theirsFile); err != nil {
		return fmt.Errorf("cannot read HEAD:%s — modify/delete conflict", path)
	}

	base, err := ParseFile(baseFile)
	if err != nil {
		return fmt.Errorf("parsing base: %w", err)
	}
	ours, err := ParseFile(oursFile)
	if err != nil {
		return fmt.Errorf("parsing ours: %w", err)
	}
	theirs, err := ParseFile(theirsFile)
	if err != nil {
		return fmt.Errorf("parsing theirs: %w", err)
	}

	merged, err := Merge(base, ours, theirs, goCeiling)
	if err != nil {
		return err
	}

	merged.Cleanup()
	out, err := merged.Format()
	if err != nil {
		return fmt.Errorf("formatting: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}

func ParseFile(path string) (*modfile.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lax := func(_, version string) (string, error) { return version, nil }
	return modfile.Parse(path, data, lax)
}

func Merge(base, ours, theirs *modfile.File, goCeiling string) (*modfile.File, error) {
	lax := func(_, v string) (string, error) { return v, nil }
	result, err := modfile.Parse("merged", modfile.Format(ours.Syntax), lax)
	if err != nil {
		return nil, fmt.Errorf("copying ours: %w", err)
	}

	goVersion := mergeGoVersion(ours, theirs, goCeiling)
	if err := result.AddGoStmt(goVersion); err != nil {
		return nil, fmt.Errorf("setting go version: %w", err)
	}

	if err := mergeRequires(base, ours, theirs, result); err != nil {
		return nil, fmt.Errorf("merging requires: %w", err)
	}

	mergeReplaces(base, ours, theirs, result)

	return result, nil
}

func mergeGoVersion(ours, theirs *modfile.File, ceiling string) string {
	oursVer := ""
	if ours.Go != nil {
		oursVer = ours.Go.Version
	}
	theirsVer := ""
	if theirs.Go != nil {
		theirsVer = theirs.Go.Version
	}
	return semver.ClampToMax(semver.Newer(oursVer, theirsVer), ceiling)
}

func mergeRequires(base, ours, theirs, result *modfile.File) error {
	baseReqs := requireMap(base)
	oursReqs := requireMap(ours)
	theirsReqs := requireMap(theirs)

	allPaths := make(map[string]bool)
	for p := range oursReqs {
		allPaths[p] = true
	}
	for p := range theirsReqs {
		allPaths[p] = true
	}

	sorted := make([]string, 0, len(allPaths))
	for p := range allPaths {
		sorted = append(sorted, p)
	}
	sort.Strings(sorted)

	indirectFlags := make(map[string]bool)

	for _, path := range sorted {
		baseReq := baseReqs[path]
		oursReq := oursReqs[path]
		theirsReq := theirsReqs[path]

		baseVer := ""
		if baseReq != nil {
			baseVer = baseReq.Mod.Version
		}

		chosen := ResolveVersion(baseVer, oursReq, theirsReq)
		if chosen == nil {
			_ = result.DropRequire(path)
			continue
		}

		if err := result.AddRequire(chosen.Mod.Path, chosen.Mod.Version); err != nil {
			return fmt.Errorf("adding require %s: %w", path, err)
		}
		indirectFlags[path] = chosen.Indirect
	}

	for _, r := range result.Require {
		if indirect, ok := indirectFlags[r.Mod.Path]; ok {
			r.Indirect = indirect
		}
	}
	result.SetRequireSeparateIndirect(result.Require)

	return nil
}

func ResolveVersion(baseVer string, ours, theirs *modfile.Require) *modfile.Require {
	if ours == nil && theirs == nil {
		return nil
	}
	if ours == nil {
		if theirs.Mod.Version == baseVer {
			return nil
		}
		return theirs
	}
	if theirs == nil {
		if ours.Mod.Version == baseVer {
			return nil
		}
		return ours
	}

	if ours.Mod.Version == theirs.Mod.Version {
		return ours
	}

	oursChanged := ours.Mod.Version != baseVer
	theirsChanged := theirs.Mod.Version != baseVer

	if oursChanged && !theirsChanged {
		return ours
	}
	if !oursChanged && theirsChanged {
		return theirs
	}

	if semver.Compare(ours.Mod.Version, theirs.Mod.Version) >= 0 {
		return ours
	}
	return theirs
}

func mergeReplaces(base, ours, theirs, result *modfile.File) {
	baseReplaces := replaceMap(base)
	oursReplaces := replaceMap(ours)
	theirsReplaces := replaceMap(theirs)

	for path, rep := range theirsReplaces {
		if _, exists := oursReplaces[path]; !exists {
			result.AddReplace(rep.Old.Path, rep.Old.Version, rep.New.Path, rep.New.Version)
		}
	}

	for path := range oursReplaces {
		_, inBase := baseReplaces[path]
		_, inTheirs := theirsReplaces[path]
		if inBase && !inTheirs {
			result.DropReplace(path, "")
		}
	}
}

func requireMap(f *modfile.File) map[string]*modfile.Require {
	m := make(map[string]*modfile.Require, len(f.Require))
	for _, r := range f.Require {
		m[r.Mod.Path] = r
	}
	return m
}

func replaceMap(f *modfile.File) map[string]*modfile.Replace {
	m := make(map[string]*modfile.Replace, len(f.Replace))
	for _, r := range f.Replace {
		m[r.Old.Path] = r
	}
	return m
}
