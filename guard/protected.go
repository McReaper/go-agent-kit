package guard

import (
	"path"
	"regexp"
	"strings"
)

var (
	testHook = regexp.MustCompile(`^func (TestMain\(|init\(\))`)
	goLine   = regexp.MustCompile(`^(go|toolchain) \S+$`)
)

func protected(f File) []Finding {
	var out []Finding
	if isTestFile(f.Path) && f.Status != 'A' {
		out = append(out, Finding{
			"protected-test", f.Path, 0,
			"T2: a test changed: the owner approves it before merge; say in your report what changed in it and why",
		})
	}
	if isTestFile(f.Path) && f.Status == 'A' {
		out = append(out, newTestHooks(f)...)
	}
	if path.Base(f.Path) == "go.mod" {
		out = append(out, goLineChanges(f)...)
	}
	return out
}

func isTestFile(p string) bool {
	return strings.HasSuffix(p, "_test.go") || strings.Contains("/"+p, "/testdata/")
}

func newTestHooks(f File) []Finding {
	var out []Finding
	for _, l := range f.Added {
		if testHook.MatchString(l.Text) {
			out = append(out, Finding{
				"protected-testmain", f.Path, l.N,
				"T2: a new TestMain or init in a test runs before every test of the package and can skip them all: the owner approves it before merge",
			})
		}
	}
	return out
}

func goLineChanges(f File) []Finding {
	var out []Finding
	for _, l := range f.Added {
		if goLine.MatchString(strings.TrimSpace(l.Text)) {
			out = append(out, Finding{
				"protected-gomod", f.Path, l.N,
				"M8: the go and toolchain lines choose the compiler and the language: the owner approves a move before merge; a dependency bump must not move them by itself",
			})
		}
	}
	return out
}
