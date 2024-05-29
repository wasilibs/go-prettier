package prettier

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/wasilibs/go-prettier/internal/runner"
)

//go:embed testdata/in
var testFiles embed.FS

//go:embed testdata/exp
var expFiles embed.FS

//go:embed testdata/exptabwidth4
var expFilesTabWidth4 embed.FS

func TestRun(t *testing.T) {
	t.Parallel()

	testFiles, _ := fs.Sub(testFiles, "testdata/in")
	expFiles, _ := fs.Sub(expFiles, "testdata/exp")
	expFilesTabWidth4, _ := fs.Sub(expFilesTabWidth4, "testdata/exptabwidth4")

	tests := []struct {
		name  string
		args  runner.RunArgs
		expFS fs.FS
	}{
		{
			name: "no config, write",
			args: runner.RunArgs{
				Write: true,
			},
			expFS: expFiles,
		},
		{
			name: "json config, write",
			args: runner.RunArgs{
				Write:  true,
				Config: filepath.Join("testdata", "config", ".prettierrc"),
			},
			expFS: expFilesTabWidth4,
		},
		{
			name: "yaml config, write",
			args: runner.RunArgs{
				Write:  true,
				Config: filepath.Join("testdata", "config", "prettierrc.yaml"),
			},
			expFS: expFilesTabWidth4,
		},
		{
			name: "toml config, write",
			args: runner.RunArgs{
				Write:  true,
				Config: filepath.Join("testdata", "config", "prettierrc.toml"),
			},
			expFS: expFilesTabWidth4,
		},
	}

	r := runner.NewRunner()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			if err := fs.WalkDir(testFiles, ".", func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() {
					return nil
				}

				c, _ := fs.ReadFile(testFiles, path)
				if err := os.WriteFile(filepath.Join(dir, path), c, 0o644); err != nil {
					return fmt.Errorf("failed to write to temp dir: %w", err)
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}

			args := tc.args
			args.Patterns = append(args.Patterns, dir)
			if err := r.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}

			if err := fs.WalkDir(tc.expFS, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				got, err := os.ReadFile(filepath.Join(dir, path))
				if err != nil {
					return fmt.Errorf("failed to read from temp dir: %w", err)
				}

				want, _ := fs.ReadFile(tc.expFS, path)
				if string(got) != string(want) {
					t.Errorf("%s - got: %s, want: %s", path, got, want)
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
