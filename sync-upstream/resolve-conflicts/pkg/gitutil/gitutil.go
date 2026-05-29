package gitutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ConflictingFiles() ([]string, error) {
	out, err := exec.Command("git", "diff", "--name-only", "--diff-filter=U").Output()
	if err != nil {
		return nil, fmt.Errorf("listing conflicting files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func MergeBase() string {
	out, err := exec.Command("git", "merge-base", "HEAD", "MERGE_HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ShowToFile(ref, outPath string) error {
	out, err := exec.Command("git", "show", ref).Output()
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, out, 0644)
}

func Add(path string) error {
	if err := exec.Command("git", "add", path).Run(); err != nil {
		return fmt.Errorf("git add %s: %w", path, err)
	}
	return nil
}
