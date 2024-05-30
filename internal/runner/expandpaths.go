package runner

import (
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

func expandPatterns(ctx context.Context, args RunArgs) []expandedPath {
	var res []expandedPath

	var expanded []expandedPattern

	customIgnore := `
.git
.sl
.svn
.hg
`
	if !args.WithNodeModules {
		customIgnore += `
node_modules
`
	}

	for _, pattern := range args.Patterns {
		pattern = filepath.Join(args.Cwd, pattern)
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
			customIgnore += filepath.ToSlash(pattern[1:]) + "\n"
		default:
			expanded = append(expanded, expandedPattern{pathType: pathTypeGlob, path: pattern})
		}
	}

	base, _ := filepath.Abs(args.Cwd)

	var ignores []gitignore.GitIgnore
	for _, p := range args.IgnorePaths {
		// Unlike upstream, we try to match git behavior better by
		// finding all .gitignore files in the repository. Notably,
		// this will find the root one when working in a subdirectory.
		if p == ".gitignore" {
			dir := base
			for {
				ignore, err := gitignore.NewFromFile(filepath.Join(dir, ".gitignore"))
				if err == nil {
					ignores = append(ignores, ignore)
				}

				if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
					break
				}

				parent := filepath.Dir(dir)
				if parent == dir || parent == "" {
					break
				}

				dir = parent
			}

			continue
		}

		ignore, err := gitignore.NewFromFile(p)
		if err != nil {
			// Prettier ignores missing ignores
			continue
		}
		ignores = append(ignores, ignore)
	}

	ignores = append(ignores, gitignore.New(strings.NewReader(customIgnore), base, nil))

	seen := map[string]struct{}{}
PathsLoop:
	for _, ep := range expanded {
		switch ep.pathType {
		case pathTypeFile:
			for _, ignore := range ignores {
				if ignore.Ignore(ep.path) {
					continue PathsLoop
				}
			}

			if _, ok := seen[ep.path]; !ok {
				res = append(res, expandedPath{filePath: ep.path})
				seen[ep.path] = struct{}{}
			}
		case pathTypeDir:
			if err := filepath.WalkDir(ep.path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				p, _ := filepath.Abs(path)
				if p == base {
					return nil
				}
				for _, ignore := range ignores {
					if m := ignore.Absolute(p, d.IsDir()); m != nil && m.Ignore() {
						if d.IsDir() {
							return filepath.SkipDir
						} else {
							return nil
						}
					}
				}

				if d.IsDir() {
					return nil
				}

				if _, ok := seen[path]; !ok {
					res = append(res, expandedPath{filePath: path, ignoreUnknown: true})
					seen[path] = struct{}{}
				}

				return nil
			}); err != nil {
				res = append(res, expandedPath{error: fmt.Sprintf(`Unable to expand directory: "%s".\n%s`, ep.path, err)})
			}
		case pathTypeGlob:
			matched := false
			if err := doublestar.GlobWalk(os.DirFS(args.Cwd), ep.path, func(path string, d fs.DirEntry) error {
				p, _ := filepath.Abs(path)
				if p == base {
					return nil
				}
				for _, ignore := range ignores {
					if m := ignore.Absolute(p, d.IsDir()); m != nil && m.Ignore() {
						if d.IsDir() {
							return filepath.SkipDir
						} else {
							return nil
						}
					}
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
