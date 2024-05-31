package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"

	"github.com/wasilibs/go-prettier/internal/runner"
)

const usage = `
Usage: prettier [options] [file/dir/glob ...]

By default, output is written to stdout.

Output options:

  -c, --check              Check if the given files are formatted, print a human-friendly summary
                           message and paths to unformatted files (see also --list-different).
  -w, --write              Edit files in-place. (Beware!)

Config options:

  --config <path>          Path to a Prettier configuration file (.prettierrc, package.json, prettier.config.js).
  --no-config              Do not look for a configuration file.
  --ignore-path <path>     Path to a file with patterns describing files to ignore.
                           Multiple values are accepted.
                           Defaults to [.gitignore, .prettierignore].
  --with-node-modules      Process files inside 'node_modules' directory.

Other options:

  --no-color               Do not colorize error messages.
  --no-error-on-unmatched-pattern
                           Prevent errors when pattern is unmatched.
  -h, --help               Show CLI usage
  -u, --ignore-unknown     Ignore unknown files.
  --log-level <silent|error|warn|log|debug>
                           What level of logs to report.
                           Defaults to log.
`

func main() {
	slog.SetDefault(slog.New(handler{level: slog.LevelInfo}))

	var args runner.RunArgs

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", strings.TrimSpace(usage))
	}

	flag.BoolVar(&args.Check, "check", false, "Check if the given files are formatted, print a human-friendly summary message and paths to unformatted files")
	flag.BoolVar(&args.Check, "c", false, "Check if the given files are formatted, print a human-friendly summary message and paths to unformatted files")
	flag.BoolVar(&args.Write, "write", false, "Edit files in-place. (Beware!)")
	flag.BoolVar(&args.Write, "w", false, "Edit files in-place. (Beware!)")

	var ignorePaths sliceFlag
	flag.Var(&ignorePaths, "ignore-path", "Path to a file with patterns describing files to ignore.\nMultiple values are accepted.\nDefaults to [.gitignore, .prettierignore].")

	flag.BoolVar(&args.NoConfig, "no-config", false, "Do not look for a configuration file.")
	flag.BoolVar(&args.NoErrorOnUnmatchedPattern, "no-error-on-unmatched-pattern", false, "Prevent errors when pattern is unmatched.")
	flag.BoolVar(&args.WithNodeModules, "with-node-modules", false, "Process files inside 'node_modules' directory.")

	levelFlg := flag.String("log-level", "log", "<silent|error|warn|log|debug>\nWhat level of logs to report.\nDefaults to log.")

	flag.Parse()

	var level slog.Level
	switch strings.ToLower(*levelFlg) {
	case "silent":
		level = slog.Level(math.MaxInt)
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "log":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	default:
		printInvalidEnumFlagValue("log-level", *levelFlg, "debug", "error", "log", "silent", "warn")
		os.Exit(1)
	}
	slog.SetDefault(slog.New(handler{level: level}))

	args.Cwd = "."
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

func printInvalidEnumFlagValue(flag string, value string, choices ...string) {
	slog.Error(fmt.Sprintf(`Invalid %s value. Expected %s, but received %s.`, colorize(red, "--"+flag), colorize(blue, "one of the following values"), colorize(red, fmt.Sprintf(`"%s"`, value))))
	for _, choice := range choices {
		slog.Error("- " + colorize(blue, fmt.Sprintf(`"%s"`, choice)))
	}
}
