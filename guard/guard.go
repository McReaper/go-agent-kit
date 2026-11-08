package guard

import (
	"context"
	"go/build/constraint"
	"path"
	"regexp"
	"slices"
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

// Check returns every rule d breaks that its lines show: bypasses, which
// block everywhere, and protected changes and size, which need a code
// owner's approval before merge (see NeedsApproval). A key or an env file
// committed is gitleaks', with Rumblestrip's rules.
func Check(d Diff) []Finding {
	var out []Finding
	for _, f := range d.Files {
		out = append(out, bypasses(f)...)
		out = append(out, protected(f)...)
	}
	return append(out, size(d)...)
}

// Run reads the change from base to the working tree of the git repository
// at dir and returns it with every rule it breaks: Check's, and what it does
// to the spec and to the libraries' API, which need the repository to read.
func Run(ctx context.Context, dir, base string) (Diff, []Finding, error) {
	d, err := Read(ctx, dir, base)
	if err != nil {
		return Diff{}, nil, err
	}
	out := Check(d)
	specs, err := specChanges(ctx, dir, base)
	if err != nil {
		return Diff{}, nil, err
	}
	api, err := apiChanges(ctx, dir, d)
	if err != nil {
		return Diff{}, nil, err
	}
	return d, append(append(out, specs...), api...), nil
}

// Subject is what a finding is about, in the owner's words: its message
// without the criterion before it and the instruction to the agent after
// it, a message being "criterion: subject: instruction".
func (f Finding) Subject() string {
	_, rest, _ := strings.Cut(f.Message, ": ")
	if i := strings.LastIndex(rest, ": "); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// NeedsApproval reports whether f is a change the owner may approve rather
// than a rule broken: the spec, the harness, a dependency or the go line, a
// library's API, or a change over the size.
func (f Finding) NeedsApproval() bool {
	return strings.HasPrefix(f.Rule, "protected-") || strings.HasPrefix(f.Rule, "size-")
}

// Blocking returns the findings that fail the check. Locally, only the
// broken rules do: protected changes are reported for the owner. In CI
// they block too, unless a code owner's approval covers them (see Covered).
func Blocking(fs []Finding, ci bool, covered []Finding) []Finding {
	var out []Finding
	for _, f := range fs {
		if !f.NeedsApproval() || ci && !slices.Contains(covered, f) {
			out = append(out, f)
		}
	}
	return out
}

var (
	suppression = regexp.MustCompile(`//\s*nolint\b|#nosec\b|//lint:(file-)?ignore\b`)
	hidingDir   = regexp.MustCompile(`(^|/)[_.][^/]*/`)
	generated   = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)
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
	if isTestFile(f.Path) {
		out = append(out, testHooks(f)...)
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
