package cmd

import (
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/fips"
)

func Fips(args []string) {
	resolved, partial, failed, err := fips.ResolveAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fips: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n[fips] Resolved: %d, Partial: %d, Failed: %d\n", len(resolved), len(partial), len(failed))
	if len(failed) > 0 {
		fmt.Println("Failed (resolve manually):")
		for _, f := range failed {
			fmt.Printf("  %s\n", f)
		}
	}
}
