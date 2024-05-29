package runner

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/denormal/go-gitignore"
)

// https://github.com/prettier/prettier/blob/main/src/cli/expand-patterns.js

type pathType byte

const (
	pathTypeFile pathType = iota
	pathTypeDir
	pathTypeGlob
)

type expandedPath struct {
	filePath      string
	ignoreUnknown bool
	error         string
}

type expandedPattern struct {
	pathType pathType
	path     string
}

func expandPatterns(ctx context.Context, args RunArgs, root string) []expandedPath {
	var res []expandedPath

	var expanded []expandedPattern

	ignoreFile := `
.git
.sl
.svn
.hg
`
	if !args.WithNodeModules {
		ignoreFile += `
node_modules
`
	}

	for _, p := range args.IgnorePaths {
		f, err := os.Open(filepath.Join(root, p))
		if err != nil {
			continue
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			ignoreFile += s.Text() + "\n"
		}
	}

	for _, pattern := range args.Patterns {
		fi, err := os.Lstat(pattern)
		switch {
		case err == nil:
			switch {
			case fi.Mode()&os.ModeSymlink != 0:
				if args.NoErrorOnUnmatchedPattern {
					res = append(res, expandedPath{error: fmt.Sprintf(`Explicitly specified pattern "%s" is a symbolic link.`, pattern)})
				} else {
					slog.DebugContext(ctx, fmt.Sprintf(`Skipping pattern "%s", as it is a symbolic link.`, pattern))
				}
			case fi.Mode().IsRegular():
				expanded = append(expanded, expandedPattern{pathType: pathTypeFile, path: pattern})
			case fi.Mode().IsDir():
				expanded = append(expanded, expandedPattern{pathType: pathTypeDir, path: pattern})
			}
		case pattern[0] == '!':
			ignoreFile += filepath.ToSlash(pattern[1:]) + "\n"
		default:
			expanded = append(expanded, expandedPattern{pathType: pathTypeGlob, path: pattern})
		}
	}

	base, _ := filepath.Abs(root)
	ignore := gitignore.New(strings.NewReader(ignoreFile), base, nil)

	seen := map[string]struct{}{}
	for _, ep := range expanded {
		switch ep.pathType {
		case pathTypeFile:
			if ignore.Ignore(ep.path) {
				continue
			}

			if _, ok := seen[ep.path]; !ok {
				res = append(res, expandedPath{filePath: ep.path})
				seen[ep.path] = struct{}{}
			}
		case pathTypeDir:
			if err := filepath.Walk(ep.path, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				p, _ := filepath.Abs(path)
				if p == base {
					return nil
				}
				if m := ignore.Absolute(p, fi.IsDir()); m != nil && m.Ignore() {
					return filepath.SkipDir
				}

				if fi.IsDir() {
					return nil
				}

				if _, ok := seen[path]; !ok {
					res = append(res, expandedPath{filePath: path, ignoreUnknown: true})
				}

				return nil
			}); err != nil {
				res = append(res, expandedPath{error: fmt.Sprintf(`Unable to expand directory: "%s".\n%s`, ep.path, err)})
			}
		case pathTypeGlob:
			matched := false
			if err := doublestar.GlobWalk(os.DirFS("."), ep.path, func(path string, d fs.DirEntry) error {
				p, _ := filepath.Abs(path)
				if p == base {
					return nil
				}
				if m := ignore.Absolute(p, d.IsDir()); m != nil && m.Ignore() {
					return filepath.SkipDir
				}

				if d.IsDir() {
					return nil
				}

				matched = true
				if _, ok := seen[path]; !ok {
					res = append(res, expandedPath{filePath: path})
					seen[path] = struct{}{}
				}

				return nil
			}, doublestar.WithNoFollow()); err != nil {
				res = append(res, expandedPath{error: fmt.Sprintf(`Unable to expand glob pattern: "%s".\n%s`, ep.path, err)})
			}
			if !matched && !args.NoErrorOnUnmatchedPattern {
				res = append(res, expandedPath{error: fmt.Sprintf(`No files matching the pattern were found: "%s".`, ep.path)})
			}
		}
	}

	return res
}
