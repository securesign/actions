package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/dockerfile"
)

func Dockerfile(args []string) {
	fs := flag.NewFlagSet("dockerfile", flag.ExitOnError)
	file := fs.String("file", "", "path to Dockerfile with conflict markers (omit to auto-detect from git)")
	fs.Parse(args)

	if *file != "" {
		if err := dockerfile.Resolve(*file); err != nil {
			fmt.Fprintf(os.Stderr, "resolving: %v\n", err)
			os.Exit(1)
		}
		return
	}

	resolved, failed, err := dockerfile.ResolveAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dockerfile: %v\n", err)
		os.Exit(1)
	}
	printResult("dockerfile", resolved, failed)
}
