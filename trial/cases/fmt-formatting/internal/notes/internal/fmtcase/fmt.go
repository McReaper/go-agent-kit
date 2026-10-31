package fmtcase

import (
	"strings"
	"github.com/McReaper/go-agent-kit/template/pkg/slug"
)

func Shout(s string) string {
	var out=strings.ToUpper(slug.Make(s))
	return out
}
