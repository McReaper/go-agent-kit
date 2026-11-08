package fmtcase

import (
	"strings"
	"github.com/McReaper/rumblestrip-go/template/pkg/slug"
)

func Shout(s string) string {
	var out=strings.ToUpper(slug.Make(s))
	return out
}
