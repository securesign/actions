package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/dockerfile"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/fips"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/gomod"
	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/workflow"
)

func Run(args []string) {
	fs := flag.NewFlagSet("all", flag.ExitOnError)
	goCeiling := fs.String("go-ceiling", "", "maximum allowed Go version (optional)")
	fs.Parse(args)

	var allResolved, allFailed []string

	fr, _, ff, err := fips.ResolveAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fips: %v\n", err)
		os.Exit(1)
	}
	allResolved = append(allResolved, fr...)
	allFailed = append(allFailed, ff...)

	r, f, err := dockerfile.ResolveAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dockerfile: %v\n", err)
		os.Exit(1)
	}
	allResolved = append(allResolved, r...)
	allFailed = append(allFailed, f...)

	r, f, err = gomod.ResolveAll(*goCeiling)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gomod: %v\n", err)
		os.Exit(1)
	}
	allResolved = append(allResolved, r...)
	allFailed = append(allFailed, f...)

	r, f, err = workflow.ResolveAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "workflow: %v\n", err)
		os.Exit(1)
	}
	allResolved = append(allResolved, r...)
	allFailed = append(allFailed, f...)

	printResult("all", allResolved, allFailed)
}

func printResult(name string, resolved, failed []string) {
	fmt.Printf("\n[%s] Resolved: %d, Failed: %d\n", name, len(resolved), len(failed))
	if len(failed) > 0 {
		fmt.Println("Failed (resolve manually):")
		for _, f := range failed {
			fmt.Printf("  %s\n", f)
		}
	}
}
