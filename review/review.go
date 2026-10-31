// Package review writes a change's review sheet: what the owner reads to
// judge it without reading the diff. It lists the changes that need their
// approval, the test scenarios added, removed or changed, read as
// sentences, contract and API changes, new modules and dependencies, shared
// code and rules touched, and the decisions (// WHY:) the change records.
package review

import (
	"context"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/internal/platform/git"
)

// Item is one line of the sheet: a file, a line when it has one, and what
// to read there.
type Item struct {
	File string
	Line int
	Text string
}

// Sheet is a change's review sheet.
type Sheet struct {
	Approval []Item // protected changes and size, from the guard
	Tests    Tests
	Contract []Item // golden files changed
	API      []Item // a library's exported API, from apidiff
	Modules  []string
	Deps     []Item
	Shared   []Item // code other modules or the app build on
	Rules    []Item // CI, the agent's settings and instructions, decisions
	Why      []Item
}

// Build reads the change from base's merge base to the working tree of the
// git repository at dir, the module's root, and writes its sheet.
func Build(ctx context.Context, dir, base string) (Sheet, error) {
	d, err := guard.Read(ctx, dir, base)
	if err != nil {
		return Sheet{}, err
	}
	s := Sheet{Approval: approval(d), Contract: goldens(d), Deps: deps(d), Shared: shared(d), Rules: rules(d), Why: decisions(d)}
	if s.Tests, err = scenarios(ctx, dir, d); err != nil {
		return Sheet{}, err
	}
	if s.Modules, err = newModules(ctx, dir, d); err != nil {
		return Sheet{}, err
	}
	if s.API, err = apiChanges(ctx, dir, d); err != nil {
		return Sheet{}, err
	}
	return s, nil
}

// Labels are the pull request labels the kit sets, each with whether this
// sheet earns it: the caller adds the ones that are true and removes the
// others.
func (s Sheet) Labels() map[string]bool {
	return map[string]bool{
		"needs-approval": len(s.Approval) > 0,
		"tests":          !s.Tests.empty(),
		"contract":       len(s.Contract)+len(s.API) > 0,
		"new-module":     len(s.Modules) > 0,
		"dependencies":   len(s.Deps) > 0,
		"shared-code":    len(s.Shared) > 0,
		"rules":          len(s.Rules) > 0,
	}
}

func approval(d guard.Diff) []Item {
	var out []Item
	for _, f := range guard.Check(d) {
		if f.NeedsApproval() {
			out = append(out, Item{File: f.File, Line: f.Line, Text: f.Message})
		}
	}
	return out
}

func goldens(d guard.Diff) []Item {
	var out []Item
	for _, f := range d.Files {
		if strings.Contains("/"+f.Path, "/testdata/") && strings.Contains(path.Base(f.Path), ".golden") {
			out = append(out, Item{File: f.Path, Text: statusWord(f.Status)})
		}
	}
	return out
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

var requirement = regexp.MustCompile(`^\s*(require\s+|replace\s+|exclude\s+|retract\s+)?[\w.~/-]+(\s+v\S+)?(\s+=>.*)?(\s*//.*)?$`)

func deps(d guard.Diff) []Item {
	var out []Item
	for _, f := range d.Files {
		if path.Base(f.Path) == "go.mod" {
			out = append(out, depsIn(f)...)
		}
	}
	return out
}

// WHY: the go and toolchain lines are the guard's (protected-gomod); the
// sheet lists what the module depends on.
func depsIn(f guard.File) []Item {
	var out []Item
	for _, l := range f.Added {
		t := strings.TrimSpace(l.Text)
		if isDependency(t) {
			out = append(out, Item{File: f.Path, Line: l.N, Text: t})
		}
	}
	return out
}

func isDependency(line string) bool {
	for _, p := range []string{"go ", "toolchain ", "module "} {
		if strings.HasPrefix(line, p) {
			return false
		}
	}
	return strings.Contains(line, " ") && requirement.MatchString(line)
}

func shared(d guard.Diff) []Item {
	var out []Item
	for _, f := range d.Files {
		if path.Ext(f.Path) != ".go" || strings.HasSuffix(f.Path, "_test.go") {
			continue
		}
		if kind := sharedKind(f.Path); kind != "" {
			out = append(out, Item{File: f.Path, Text: kind + ", " + statusWord(f.Status)})
		}
	}
	return out
}

func sharedKind(p string) string {
	switch {
	case strings.HasPrefix(p, "internal/platform/"):
		return "platform"
	case strings.HasPrefix(p, "pkg/"):
		return "library"
	case strings.HasPrefix(p, "cmd/"):
		return "composition root"
	}
	if dir, m, ok := guard.Module(p); ok && m != "platform" && path.Dir(p) == dir {
		return "the " + m + " module's contract"
	}
	return ""
}

func rules(d guard.Diff) []Item {
	var out []Item
	for _, f := range d.Files {
		if kind := rulesKind(f.Path); kind != "" {
			out = append(out, Item{File: f.Path, Text: kind + ", " + statusWord(f.Status)})
		}
	}
	return out
}

func rulesKind(p string) string {
	base := path.Base(p)
	switch {
	case strings.HasPrefix(p, ".github/"):
		return "CI and code owners"
	case strings.HasPrefix(p, ".claude/") || strings.HasPrefix(p, ".kit/"):
		return "the agent's settings"
	case base == "AGENTS.md" || base == "CLAUDE.md":
		return "the agent's instructions"
	case strings.HasPrefix(p, "docs/adr/"):
		return "a decision record"
	case p == "renovate.json":
		return "the dependency bot's rules"
	}
	return ""
}

var (
	whyComment  = regexp.MustCompile(`^\s*// WHY: ?(.*)$`)
	commentLine = regexp.MustCompile(`^\s*// ?(.*)$`)
)

// WHY: a decision often spans lines; the sheet shows it whole, with the
// comment lines that continue it.
func decisions(d guard.Diff) []Item {
	var out []Item
	for _, f := range d.Files {
		if path.Ext(f.Path) != ".go" {
			continue
		}
		out = append(out, decisionsIn(f)...)
	}
	return out
}

func decisionsIn(f guard.File) []Item {
	var out []Item
	for i := 0; i < len(f.Added); i++ {
		m := whyComment.FindStringSubmatch(f.Added[i].Text)
		if m == nil {
			continue
		}
		text := []string{m[1]}
		for i+1 < len(f.Added) && f.Added[i+1].N == f.Added[i].N+1 && continuesComment(f.Added[i+1].Text) {
			i++
			text = append(text, commentLine.FindStringSubmatch(f.Added[i].Text)[1])
		}
		out = append(out, Item{File: f.Path, Line: f.Added[i-len(text)+1].N, Text: strings.Join(text, " ")})
	}
	return out
}

func continuesComment(line string) bool {
	return commentLine.MatchString(line) && !whyComment.MatchString(line)
}

func newModules(ctx context.Context, dir string, d guard.Diff) ([]string, error) {
	var out []string
	for _, f := range d.Files {
		mdir, name, ok := guard.Module(f.Path)
		if !ok || name == "platform" || f.Status == 'D' || slices.Contains(out, mdir) {
			continue
		}
		at, err := git.Run(ctx, dir, "ls-tree", d.Base, "--", mdir)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(at) == "" {
			out = append(out, mdir)
		}
	}
	return out, nil
}
