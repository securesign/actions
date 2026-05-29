package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/securesign/actions/sync-upstream/resolve-conflicts/pkg/pyxis"
)

func GoCeiling(args []string) {
	fs := flag.NewFlagSet("go-ceiling", flag.ExitOnError)
	image := fs.String("image", "registry.redhat.io/ubi9/go-toolset:latest", "container image to check for Go version")
	fs.Parse(args)

	version, err := pyxis.GetGoVersion(*image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(version)
}
