package guard_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/guard"
)

func TestThePlatformIsNoModule(t *testing.T) {
	for p, want := range map[string]string{
		"internal/notes/notes.go":                   "notes",
		"internal/notes/internal/add/add.go":        "notes",
		"apps/x/internal/tags/tags.go":              "tags",
		"internal/platform/op/op.go":                "",
		"internal/platform/journal/journal_test.go": "",
		"cmd/notes/main.go":                         "",
	} {
		_, got, ok := guard.Module(p)
		if got != want || ok != (want != "") {
			t.Errorf("Module(%q) = %q, %v; want %q", p, got, ok, want)
		}
	}
}

// WHY: a feature often needs the shared platform too (a new flag type for
// its arguments); counting the platform as a second module asked the
// owner to approve the split of a change that is one.
func TestAModuleAndThePlatformAreOneModuleChange(t *testing.T) {
	d := guard.Diff{Files: []guard.File{
		{Path: "internal/notes/internal/tags/tags.go", Status: 'A', Added: []guard.Line{{N: 1, Text: "package tags"}}},
		{Path: "internal/platform/op/op.go", Status: 'M', Added: []guard.Line{{N: 9, Text: "case *[]string:"}}},
	}}
	for _, f := range guard.Check(d) {
		if f.Rule == "size-modules" {
			t.Errorf("a module and the platform: %s", f.Message)
		}
	}
	d.Files = append(d.Files, guard.File{Path: "internal/tags/tags.go", Status: 'A', Added: []guard.Line{{N: 1, Text: "package tags"}}})
	if !hasRule(guard.Check(d), "size-modules") {
		t.Error("two modules passed as one")
	}
}

func hasRule(fs []guard.Finding, rule string) bool {
	for _, f := range fs {
		if f.Rule == rule {
			return true
		}
	}
	return false
}
