package runner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergePrettierConfig(t *testing.T) {
	baseCfg := map[string]any{
		"tabWidth":   5,
		"printWidth": 99,
	}

	overridesCfg := map[string]any{
		"tabWidth":   5,
		"printWidth": 99,
		"overrides": []map[string]any{
			{
				"files":        "*.js",
				"excludeFiles": "dog.js",
				"options": map[string]any{
					"tabWidth": 4,
				},
			},
			{
				"files": []string{"cat.js"},
				"options": map[string]any{
					"printWidth": 100,
				},
			},
			{
				"files":        []string{"**/testdata/*.js"},
				"excludeFiles": []string{"cat.js", "mouse.js"},
				"options": map[string]any{
					"tabWidth": 6,
				},
			},
		},
	}

	tests := []struct {
		name string
		in   map[string]any
		exp  map[string]any
		path string
	}{
		{
			name: "no overrides",
			in:   baseCfg,
			exp:  baseCfg,
			path: "bear.js",
		},
		{
			name: "normal match",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   4,
				"printWidth": 99,
			},
			path: "bear.js",
		},
		{
			name: "basename match",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   4,
				"printWidth": 99,
			},
			path: "animals/bear.js",
		},
		{
			name: "no match",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   5,
				"printWidth": 99,
			},
			path: "bear.jsx",
		},
		{
			name: "multiple matches",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   4,
				"printWidth": 100,
			},
			path: "cat.js",
		},
		{
			name: "exclude",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   5,
				"printWidth": 99,
			},
			path: "dog.js",
		},
		{
			name: "glob match",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   6,
				"printWidth": 99,
			},
			path: "animals/testdata/bear.js",
		},
		{
			name: "glob match exclude",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   4,
				"printWidth": 100,
			},
			path: "animals/testdata/cat.js",
		},
		{
			name: "glob match exclude 2",
			in:   overridesCfg,
			exp: map[string]any{
				"tabWidth":   4,
				"printWidth": 99,
			},
			path: "animals/testdata/mouse.js",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := map[string]any{}
			mergePrettierConfig(res, tc.in, tc.path)
			require.Equal(t, tc.exp, res)
		})
	}
}
