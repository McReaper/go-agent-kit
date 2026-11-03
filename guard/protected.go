package guard

import (
	"path"
	"regexp"
	"strings"
)

var (
	goLine      = regexp.MustCompile(`^(go|toolchain) \S+$`)
	requirement = regexp.MustCompile(`^(?:require\s+)?([\w.~/-]+\.[\w.~/-]+)\s+v\S+`)
)

// Harness is what tells the agent how to work and the merge check how to
// judge it: a change to it is the owner's to approve. An app's CODEOWNERS
// names the same paths, so GitHub holds them even if the check is broken.
// A path ending in / is a directory from the root; GLOSSARY.md is any
// module's.
func Harness() []string {
	return []string{".github/", ".claude/", ".kit/", "AGENTS.md", "CLAUDE.md", "docs/adr/", "renovate.json", "GLOSSARY.md"}
}

func isHarness(p string) bool {
	for _, h := range Harness() {
		switch {
		case strings.HasSuffix(h, "/") && strings.HasPrefix(p, h), p == h, h == "GLOSSARY.md" && path.Base(p) == h:
			return true
		}
	}
	return false
}

// WHY: the owner reads the spec, not the code, so a change to a test of
// the code is the machines' to judge (mutation, the spec's own proof);
// what stays theirs is what no test can judge: the spec, the harness, what
// the module depends on and the compiler it builds with.
func protected(f File) []Finding {
	var out []Finding
	if isHarness(f.Path) || f.OldPath != "" && isHarness(f.OldPath) {
		out = append(out, Finding{
			"protected-harness", f.Path, 0,
			"T2: " + harnessKind(f.Path) + " " + statusWord(f.Status) + ": the harness tells the agent how to work and the CI how to judge, so the owner approves a change to it before merge",
		})
	}
	if path.Base(f.Path) == "go.mod" && f.Status != 'D' {
		out = append(out, goModChanges(f)...)
	}
	return out
}

func harnessKind(p string) string {
	switch {
	case strings.HasPrefix(p, ".github/"):
		return "the CI or the code owners"
	case strings.HasPrefix(p, ".claude/") || strings.HasPrefix(p, ".kit/"):
		return "the agent's settings"
	case p == "AGENTS.md" || p == "CLAUDE.md":
		return "the agent's instructions"
	case strings.HasPrefix(p, "docs/adr/"):
		return "a decision record"
	case p == "renovate.json":
		return "the dependency bot's rules"
	}
	return "a module's glossary"
}

func statusWord(s byte) string {
	switch s {
	case 'A':
		return "added"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	}
	return "changed"
}

func isTestFile(p string) bool {
	return strings.HasSuffix(p, "_test.go") || strings.Contains("/"+p, "/testdata/")
}

func goModChanges(f File) []Finding {
	had := map[string]bool{}
	for _, l := range f.Gone {
		if m := requirement.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
			had[m[1]] = true
		}
	}
	var out []Finding
	for _, l := range f.Added {
		t := strings.TrimSpace(l.Text)
		if goLine.MatchString(t) {
			out = append(out, Finding{
				"protected-gomod", f.Path, l.N,
				"M8: the " + strings.Fields(t)[0] + " line moves to " + strings.Fields(t)[1] +
					": the go and toolchain lines choose the compiler and the language, so the owner approves a move before merge, and a dependency bump must not move them by itself",
			})
		}
		m := requirement.FindStringSubmatch(t)
		if m != nil && !had[m[1]] && !strings.Contains(t, "// indirect") {
			out = append(out, Finding{
				"protected-dependency", f.Path, l.N,
				"M8: a new dependency, " + m[1] + ": what the app depends on is the owner's call, bumping one it has is not",
			})
		}
	}
	return out
}
