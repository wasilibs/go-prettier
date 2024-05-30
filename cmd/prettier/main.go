package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/wasilibs/go-prettier/internal/runner"
)

func main() {
	var args runner.RunArgs

	flag.BoolVar(&args.Check, "check", false, "Check if the given files are formatted, print a human-friendly summary message and paths to unformatted files")
	flag.BoolVar(&args.Check, "c", false, "Check if the given files are formatted, print a human-friendly summary message and paths to unformatted files")
	flag.BoolVar(&args.Write, "write", false, "Edit files in-place. (Beware!)")
	flag.BoolVar(&args.Write, "w", false, "Edit files in-place. (Beware!)")

	var ignorePaths sliceFlag
	flag.Var(&ignorePaths, "ignore-path", "Path to a file with patterns describing files to ignore.\nMultiple values are accepted.\nDefaults to [.gitignore, .prettierignore].")

	flag.BoolVar(&args.NoConfig, "no-config", false, "Do not look for a configuration file.")
	flag.BoolVar(&args.NoErrorOnUnmatchedPattern, "no-error-on-unmatched-pattern", false, "Prevent errors when pattern is unmatched.")
	flag.BoolVar(&args.WithNodeModules, "with-node-modules", false, "Process files inside 'node_modules' directory.")

	flag.Parse()

	args.Patterns = flag.Args()

	if len(ignorePaths) == 0 {
		ignorePaths = append(ignorePaths, ".gitignore", ".prettierignore")
	}
	args.IgnorePaths = ignorePaths

	r := runner.NewRunner()
	if err := r.Run(context.Background(), args); err != nil {
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
