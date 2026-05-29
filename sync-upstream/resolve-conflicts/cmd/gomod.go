package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/gomod"
)

func Gomod(args []string) {
	fs := flag.NewFlagSet("gomod", flag.ExitOnError)
	baseFile := fs.String("base", "", "path to base (merge-base) go.mod")
	oursFile := fs.String("ours", "", "path to ours (downstream) go.mod")
	theirsFile := fs.String("theirs", "", "path to theirs (upstream) go.mod")
	goCeiling := fs.String("go-ceiling", "", "maximum allowed Go version")
	outputFile := fs.String("output", "", "path to write merged go.mod")
	fs.Parse(args)

	// Explicit three-way merge mode
	if *baseFile != "" || *oursFile != "" || *theirsFile != "" || *outputFile != "" {
		if *baseFile == "" || *oursFile == "" || *theirsFile == "" || *outputFile == "" {
			fmt.Fprintln(os.Stderr, "usage: resolve-conflicts gomod --base FILE --ours FILE --theirs FILE --output FILE [--go-ceiling VERSION]")
			os.Exit(1)
		}

		base, err := gomod.ParseFile(*baseFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parsing base: %v\n", err)
			os.Exit(1)
		}
		ours, err := gomod.ParseFile(*oursFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parsing ours: %v\n", err)
			os.Exit(1)
		}
		theirs, err := gomod.ParseFile(*theirsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parsing theirs: %v\n", err)
			os.Exit(1)
		}

		merged, err := gomod.Merge(base, ours, theirs, *goCeiling)
		if err != nil {
			fmt.Fprintf(os.Stderr, "merging: %v\n", err)
			os.Exit(1)
		}

		merged.Cleanup()
		out, err := merged.Format()
		if err != nil {
			fmt.Fprintf(os.Stderr, "formatting: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(*outputFile, out, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "writing: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Auto-detect mode
	resolved, failed, err := gomod.ResolveAll(*goCeiling)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gomod: %v\n", err)
		os.Exit(1)
	}
	printResult("gomod", resolved, failed)
}
