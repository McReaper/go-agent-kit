package guard

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
)

var moduleOf = regexp.MustCompile(`^((?:[^/]+/)*?internal/([^/]+))/`)

// Module returns the directory and the name of the module p belongs to:
// the first internal/<name>/ in its path, internal/platform/ excepted,
// which every module shares.
func Module(p string) (dir, name string, ok bool) {
	m := moduleOf.FindStringSubmatch(p)
	if m == nil || m[2] == "platform" {
		return "", "", false
	}
	return m[1], m[2], true
}

func size(d Diff) []Finding {
	lines := 0
	var modules []string
	for _, f := range d.Files {
		if !isProduction(f.Path) {
			continue
		}
		lines += len(f.Added) + f.Removed
		if _, m, ok := Module(f.Path); ok && !slices.Contains(modules, m) {
			modules = append(modules, m)
		}
	}
	var out []Finding
	if lines > MaxProductionLines {
		out = append(out, Finding{"size-lines", "", 0, fmt.Sprintf(
			"T3: this change touches %d production lines, the limit is %d: split it into changes a reviewer can read", lines, MaxProductionLines)})
	}
	if len(modules) > MaxModules {
		out = append(out, Finding{"size-modules", "", 0, fmt.Sprintf(
			"T3/E2: this change touches %d modules (%s): change one module per pull request", len(modules), strings.Join(modules, ", "))})
	}
	return out
}

func isProduction(p string) bool {
	return path.Ext(p) == ".go" && !isTestFile(p) && !isGeneratedPath(p)
}
