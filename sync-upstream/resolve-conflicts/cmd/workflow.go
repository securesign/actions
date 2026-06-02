package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/workflow"
)

func Workflow(args []string) {
	fs := flag.NewFlagSet("workflow", flag.ExitOnError)
	file := fs.String("file", "", "path to workflow file with conflict markers (omit to auto-detect from git)")
	fs.Parse(args)

	if *file != "" {
		if err := workflow.Resolve(*file); err != nil {
			fmt.Fprintf(os.Stderr, "resolving: %v\n", err)
			os.Exit(1)
		}
		return
	}

	resolved, failed, err := workflow.ResolveAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "workflow: %v\n", err)
		os.Exit(1)
	}
	printResult("workflow", resolved, failed)
}
