// Package slug turns free text into lowercase ASCII identifiers.
package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Make returns s in lowercase ASCII, with every run of other characters
// replaced by a single hyphen and no hyphen at either end.
func Make(s string) string {
	var b strings.Builder
	pending := false
	for _, r := range norm.NFKD.String(s) {
		switch {
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			if pending && b.Len() > 0 {
				b.WriteByte('-')
			}
			pending = false
			b.WriteRune(unicode.ToLower(r))
		case unicode.Is(unicode.Mn, r):
		default:
			pending = true
		}
	}
	return b.String()
}
