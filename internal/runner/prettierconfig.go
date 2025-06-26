package runner

import (
	"maps"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func mergePrettierConfig(mergedCfg map[string]any, userCfg map[string]any, path string) {
	var overrides []map[string]any
	for k, v := range userCfg {
		if k == "overrides" {
			if o, ok := v.([]map[string]any); ok {
				overrides = o
				continue
			}
		}
		mergedCfg[k] = v
	}

	if len(overrides) == 0 {
		return
	}

	for _, o := range overrides {
		patterns := toStrings(o["files"])
		ignores := toStrings(o["excludeFiles"])

		if matchAny(ignores, path) {
			continue
		}

		if matchAny(patterns, path) {
			if m, ok := o["options"].(map[string]any); ok {
				maps.Copy(mergedCfg, m)
			}
		}
	}
}

func toStrings(v any) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	}
	return nil
}

func matchAny(patterns []string, path string) bool {
	for _, p := range patterns {
		target := path
		// Pattern without slash is matched against basename, not full path
		if strings.IndexByte(p, '/') == -1 {
			target = filepath.Base(path)
		}

		if m, _ := doublestar.Match(p, target); m {
			return true
		}
	}
	return false
}
