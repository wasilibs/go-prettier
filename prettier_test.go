package prettier

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wasilibs/go-prettier/v3/internal/runner"
)

//go:embed testdata/in
var testFiles embed.FS

//go:embed testdata/exp
var expFiles embed.FS

//go:embed testdata/exptabwidth4
var expFilesTabWidth4 embed.FS

//go:embed testdata/config/.prettierrc
var prettierrc []byte

//go:embed testdata/config/prettierrc.yaml
var prettierrcYAML []byte

//go:embed testdata/config/prettierrc.toml
var prettierrcTOML []byte

//go:embed testdata/config/.editorconfig
var editorconfig []byte

func TestRun(t *testing.T) {
	t.Parallel()

	testFiles, _ := fs.Sub(testFiles, "testdata/in")
	expFiles, _ := fs.Sub(expFiles, "testdata/exp")
	expFilesTabWidth4, _ := fs.Sub(expFilesTabWidth4, "testdata/exptabwidth4")

	tests := []struct {
		name    string
		args    runner.RunArgs
		expFS   fs.FS
		prepare func(dir string) error
	}{
		{
			name: "no config, write",
			args: runner.RunArgs{
				Patterns: []string{"."},
				Write:    true,
			},
			expFS: expFiles,
		},
		{
			name: "json config, write",
			prepare: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".prettierrc"), prettierrc, 0o644)
			},
			args: runner.RunArgs{
				Patterns: []string{"."},
				Write:    true,
			},
			expFS: expFilesTabWidth4,
		},
		{
			name: "yaml config, write",
			prepare: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".prettierrc.yaml"), prettierrcYAML, 0o644)
			},
			args: runner.RunArgs{
				Patterns: []string{"."},
				Write:    true,
			},
			expFS: expFilesTabWidth4,
		},
		{
			name: "toml config, write",
			prepare: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".prettierrc.toml"), prettierrcTOML, 0o644)
			},
			args: runner.RunArgs{
				Patterns: []string{"."},
				Write:    true,
			},
			expFS: expFilesTabWidth4,
		},
		{
			name: "editorconfig, write",
			prepare: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".editorconfig"), editorconfig, 0o644)
			},
			args: runner.RunArgs{
				Patterns: []string{"."},
				Write:    true,
			},
			expFS: expFilesTabWidth4,
		},
	}

	r := runner.NewRunner()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			require.NoError(t, fs.WalkDir(testFiles, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				c, _ := fs.ReadFile(testFiles, path)
				if err := os.WriteFile(filepath.Join(dir, path), c, 0o644); err != nil {
					return fmt.Errorf("failed to write to temp dir: %w", err)
				}

				return nil
			}))

			if tc.prepare != nil {
				require.NoError(t, tc.prepare(dir))
			}

			args := tc.args
			args.Cwd = dir
			require.NoError(t, r.Run(t.Context(), args))

			require.NoError(t, fs.WalkDir(tc.expFS, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				have, err := os.ReadFile(filepath.Join(dir, path))
				if err != nil {
					return fmt.Errorf("failed to read from temp dir: %w", err)
				}

				exp, _ := fs.ReadFile(tc.expFS, path)
				require.Equal(t, string(exp), string(have))

				return nil
			}))
		})
	}
}
