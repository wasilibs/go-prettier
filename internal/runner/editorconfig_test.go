package runner

import (
	"math"
	"strings"
	"testing"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/stretchr/testify/require"
)

// https://github.com/prettier/prettier/blob/8a88cdce6d4605f206305ebb9204a0cabf96a070/tests/unit/editorconfig-to-prettier.js
func TestFillEditorConfig(t *testing.T) {
	tests := []struct {
		ecfg string
		exp  map[string]any
	}{
		{
			ecfg: `
			[*]
			indent_style = tab
			tab_width = 8
			indent_size = 2
			max_line_length = 100
			`,
			exp: map[string]any{
				"useTabs":    true,
				"tabWidth":   8,
				"printWidth": 100,
			},
		},
		{
			ecfg: `
			[*]
			indent_style = space
			tab_width = 8
			indent_size = 2
			max_line_length = 100
			`,
			exp: map[string]any{
				"useTabs":    false,
				"tabWidth":   2,
				"printWidth": 100,
			},
		},
		{
			ecfg: `
			[*]
			indent_style = space
			tab_width = 8
			indent_size = 8
			max_line_length = 100
			`,
			exp: map[string]any{
				"useTabs":    false,
				"tabWidth":   8,
				"printWidth": 100,
			},
		},
		{
			ecfg: `
			[*]
			tab_width = 4
			indent_size = tab
			`,
			exp: map[string]any{
				"tabWidth": 4,
				"useTabs":  true,
			},
		},
		{
			ecfg: `
			[*]
			indent_size = tab
			`,
			exp: map[string]any{
				"useTabs": true,
			},
		},
		{
			ecfg: `
			[*]
			tab_width = 0
			indent_size = 0
			`,
			exp: map[string]any{
				"tabWidth": 0,
			},
		},
		{
			ecfg: `
			[*]
			quote_type = double
			`,
			exp: map[string]any{
				"singleQuote": false,
			},
		},
		{
			ecfg: `
			[*]
			quote_type = double
			max_line_length = off
			`,
			exp: map[string]any{
				"printWidth":  math.Inf(1),
				"singleQuote": false,
			},
		},
		{
			ecfg: `
			[*]
			end_of_line = cr
			`,
			exp: map[string]any{
				"endOfLine": "cr",
			},
		},
		{
			ecfg: `
			[*]
			end_of_line = crlf
			`,
			exp: map[string]any{
				"endOfLine": "crlf",
			},
		},
		{
			ecfg: `
			[*]
			end_of_line = lf
			`,
			exp: map[string]any{
				"endOfLine": "lf",
			},
		},
		{
			ecfg: `
			[*]
			end_of_line = 123
			`,
			exp: map[string]any{},
		},
		{
			ecfg: `
			[*]
			indent_style = space
			indent_size = 2
			max_line_length = unset
			`,
			exp: map[string]any{
				"useTabs":  false,
				"tabWidth": 2,
			},
		},
		{
			ecfg: ``,
			exp:  map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			res := make(map[string]any)
			cfg, err := editorconfig.Parse(strings.NewReader(tc.ecfg))
			require.NoError(t, err)
			def, err := cfg.GetDefinitionForFilename("test")
			require.NoError(t, err)

			fillEditorConfig(def, res)
			require.Equal(t, tc.exp, res)
		})
	}
}
