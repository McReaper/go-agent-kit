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
			"T2: changing, renaming or deleting a test needs a code owner's approval: add a new test instead, or ask for review",
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
				"T2: a new TestMain or init in a test runs before every test of the package and can skip them all: it needs a code owner's approval",
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
				"M8: the go and toolchain lines choose the compiler and the language: a dependency bump must not move them; restore it or ask for review",
			})
		}
	}
	return out
}
