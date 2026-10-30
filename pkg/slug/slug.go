// Copyright 2026 Notes Inc. All rights reserved.

// Package slug turns free text into lowercase ASCII identifiers.
package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Make returns s in lowercase ASCII, with every run of other characters
// replaced by a single hyphen and no hyphen at either end. Accents are
// dropped from the letters they sit on.
func Make(s string) string {
	bare := strings.Map(dropMarks, norm.NFKD.String(s))
	return strings.ToLower(strings.Join(strings.FieldsFunc(bare, isSeparator), "-"))
}

func dropMarks(r rune) rune {
	if unicode.Is(unicode.Mn, r) {
		return -1
	}
	return r
}

func isSeparator(r rune) bool {
	return r >= unicode.MaxASCII || !unicode.IsLetter(r) && !unicode.IsDigit(r)
}
