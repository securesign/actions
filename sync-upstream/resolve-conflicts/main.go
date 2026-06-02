package main

import (
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/cmd"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "all":
		cmd.Run(os.Args[2:])
	case "gomod":
		cmd.Gomod(os.Args[2:])
	case "dockerfile":
		cmd.Dockerfile(os.Args[2:])
	case "workflow":
		cmd.Workflow(os.Args[2:])
	case "go-ceiling":
		cmd.GoCeiling(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `resolve-conflicts - upstream sync conflict resolver

Usage: resolve-conflicts <command> [flags]

Commands:
  all         Auto-resolve all known conflict patterns in the current merge
  gomod       Three-way merge of go.mod files
  dockerfile  Resolve Dockerfile golang version conflicts
  workflow    Resolve GitHub Actions workflow version conflicts
  go-ceiling  Get max Go version from Red Hat container image`)
	os.Exit(1)
}
