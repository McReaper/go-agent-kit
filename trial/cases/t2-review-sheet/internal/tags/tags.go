// Package tags is the tags module: labels a note carries.
package tags

import "strings"

// Normalize returns the tags lower-cased, in their first order, once each.
//
// WHY: tags come from people typing; "Go" and "go" are one tag, and the
// order they gave is the order they read back.
func Normalize(tags []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}
