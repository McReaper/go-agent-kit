package guard

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
)

var moduleOf = regexp.MustCompile(`^(?:[^/]+/)*?internal/([^/]+)/`)

func size(d Diff) []Finding {
	lines := 0
	var modules []string
	for _, f := range d.Files {
		if !isProduction(f.Path) {
			continue
		}
		lines += len(f.Added) + f.Removed
		if m := moduleOf.FindStringSubmatch(f.Path); m != nil && !slices.Contains(modules, m[1]) {
			modules = append(modules, m[1])
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
