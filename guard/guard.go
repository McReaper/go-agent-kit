package guard

import (
	"go/build/constraint"
	"path"
	"regexp"
	"strings"
)

// Finding is one rule a change breaks.
type Finding struct {
	Rule    string
	File    string
	Line    int
	Message string
}

// Limits a change must stay within.
const (
	MaxProductionLines = 400
	MaxModules         = 1
)

// Check returns what d breaks. Protected changes (tests, golden files, the
// go line) are findings unless approved is set: the CI sets it once the
// change's latest commit is approved by a code owner.
func Check(d Diff, approved bool) []Finding {
	var out []Finding
	for _, f := range d.Files {
		out = append(out, bypasses(f)...)
		if !approved {
			out = append(out, protected(f)...)
		}
	}
	return append(out, size(d)...)
}

var (
	suppression = regexp.MustCompile(`//\s*nolint\b|#nosec\b|//lint:(file-)?ignore\b`)
	hidingDir   = regexp.MustCompile(`(^|/)[_.][^/]*/`)
	generated   = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)
	skipCall    = regexp.MustCompile(`\.Skip(Now|f)?\(|testing\.Short\(\)`)
)

func bypasses(f File) []Finding {
	if f.Status == 'D' || path.Ext(f.Path) != ".go" {
		return nil
	}
	var out []Finding
	if hidingDir.MatchString(f.Path) {
		out = append(out, Finding{
			"bypass-hidden-dir", f.Path, 0,
			"T2: the go tool never builds, vets or tests a directory starting with _ or .: move the code where it is checked",
		})
	}
	for _, l := range f.Added {
		out = append(out, lineBypasses(f.Path, l)...)
	}
	return out
}

func lineBypasses(file string, l Line) []Finding {
	var out []Finding
	text := strings.TrimSpace(l.Text)
	add := func(rule, msg string) { out = append(out, Finding{rule, file, l.N, msg}) }
	if suppression.MatchString(text) {
		add("bypass-suppression", "T2: a finding is fixed, not silenced: remove the suppression and fix what it hid")
	}
	if strings.HasPrefix(text, "//go:linkname") || strings.HasPrefix(text, "//go:debug") || text == `import "C"` {
		add("bypass-directive", "T2: //go:linkname, //go:debug and cgo reach past the type system and the checks: remove it")
	}
	if strings.HasSuffix(file, "_test.go") && skipCall.MatchString(text) {
		add("bypass-skip", "T2: a skipped test proves nothing: make it pass, or delete it in a change a code owner approves")
	}
	if generated.MatchString(text) && !isGeneratedPath(file) {
		add("bypass-generated", "T2: the generated-code header turns linters off: only files under gen/ or named *_gen.go may carry it")
	}
	if msg, ok := customTag(text); ok {
		add("bypass-buildtag", msg)
	}
	return out
}

func isGeneratedPath(file string) bool {
	return strings.Contains("/"+file, "/gen/") || strings.HasSuffix(file, "_gen.go") || strings.HasSuffix(file, ".pb.go")
}

// WHY: a build tag no target sets hides its file from every lint and test
// run; only systems, architectures and Go versions may gate a file.
func customTag(text string) (string, bool) {
	expr, err := constraint.Parse(text)
	if err != nil {
		return "", false
	}
	var unknown []string
	expr.Eval(func(tag string) bool {
		if !knownTag(tag) {
			unknown = append(unknown, tag)
		}
		return true
	})
	if len(unknown) == 0 {
		return "", false
	}
	return "T2: build tag " + strings.Join(unknown, ", ") + " is set by no target, so the file escapes lint and tests: gate on GOOS, GOARCH or a Go version only", true
}

func knownTag(tag string) bool {
	switch tag {
	case "aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios", "js", "linux",
		"netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows", "zos", "unix",
		"386", "amd64", "arm", "arm64", "loong64", "mips", "mips64", "mips64le", "mipsle",
		"ppc64", "ppc64le", "riscv64", "s390x", "wasm", "gc":
		return true
	}
	return strings.HasPrefix(tag, "go1.")
}
