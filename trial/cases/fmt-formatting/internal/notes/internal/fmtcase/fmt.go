package fmtcase

import (
	"strings"
	"example.com/notes/pkg/slug"
)

func Shout(s string) string {
	var out=strings.ToUpper(slug.Make(s))
	return out
}
