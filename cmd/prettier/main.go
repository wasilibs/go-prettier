package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/wasilibs/go-prettier/internal/runner"
)

func main() {
	var check bool
	var write bool

	flag.BoolVar(&check, "check", false, "Check if the given files are formatted, print a human-friendly summary message and paths to unformatted files")
	flag.BoolVar(&check, "c", false, "Check if the given files are formatted, print a human-friendly summary message and paths to unformatted files")
	flag.BoolVar(&write, "write", false, "Edit files in-place. (Beware!)")
	flag.BoolVar(&write, "w", false, "Edit files in-place. (Beware!)")

	var ignorePaths sliceFlag
	flag.Var(&ignorePaths, "ignore-path", "Path to a file with patterns describing files to ignore.\nMultiple values are accepted.\nDefaults to [.gitignore, .prettierignore].")

	noConfig := flag.Bool("no-config", false, "Do not look for a configuration file.")
	noErrorOnUnmatchedPattern := flag.Bool("no-error-on-unmatched-pattern", false, "Prevent errors when pattern is unmatched.")
	withNodeModules := flag.Bool("with-node-modules", false, "Process files inside 'node_modules' directory.")

	flag.Parse()

	if len(ignorePaths) == 0 {
		ignorePaths = append(ignorePaths, ".gitignore", ".prettierignore")
	}

	r := runner.NewRunner()
	if err := r.Run(context.Background(), runner.RunArgs{
		Patterns:                  flag.Args(),
		Check:                     check,
		Write:                     write,
		IgnorePaths:               ignorePaths,
		NoConfig:                  *noConfig,
		NoErrorOnUnmatchedPattern: *noErrorOnUnmatchedPattern,
		WithNodeModules:           *withNodeModules,
	}); err != nil {
		// Runner handles logging so we just need to set error code.
		os.Exit(1)
	}
}

type sliceFlag []string

func (f *sliceFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *sliceFlag) Set(s string) error {
	*f = append(*f, s)
	return nil
}
