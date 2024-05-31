package runner

import (
	"math"
	"strconv"

	"github.com/editorconfig/editorconfig-core-go/v2"
)

// https://github.com/prettier/prettier/blob/main/src/config/editorconfig/editorconfig-to-prettier.js

func fillEditorConfig(def *editorconfig.Definition, res map[string]any) {
	if def.IndentStyle != "" {
		res["useTabs"] = def.IndentStyle == "tab"
	}

	if def.IndentSize == "tab" {
		res["useTabs"] = true
	}

	switch {
	case res["useTabs"] == true && def.TabWidth > 0:
		res["tabWidth"] = def.TabWidth
	case def.IndentStyle == "space" && def.IndentSize != "" && def.IndentSize != "tab":
		if n, err := strconv.Atoi(def.IndentSize); err == nil {
			res["tabWidth"] = n
		}
	case def.Raw["tab_width"] != "":
		res["tabWidth"] = def.TabWidth
	}

	if maxLenStr := def.Raw["max_line_length"]; isSet(maxLenStr) {
		if maxLenStr == "off" {
			res["printWidth"] = math.Inf(1)
		} else if maxLen, err := strconv.Atoi(maxLenStr); err == nil {
			res["printWidth"] = maxLen
		}
	}

	if quoteType := def.Raw["quote_type"]; isSet(quoteType) {
		switch quoteType {
		case "single":
			res["singleQuote"] = true
		case "double":
			res["singleQuote"] = false
		}
	}

	switch def.EndOfLine {
	case "cr", "crlf", "lf":
		res["endOfLine"] = def.EndOfLine
	}
}

func isSet(val string) bool {
	return val != "" && val != editorconfig.UnsetValue
}
