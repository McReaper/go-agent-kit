package guard_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/google/go-cmp/cmp"
)

func TestParseReadsStatusesRenamesAndAddedLines(t *testing.T) {
	status := "M\ta.go\nR087\told/b.go\tnew/b.go\nD\tc_test.go\n"
	numstat := "2\t1\ta.go\n1\t3\t{old => new}/b.go\n0\t9\tc_test.go\n"
	patch := "+++ a.go\n@@ -3,1 +3,2 @@\n+x := 1\n+y := 2\n+++ new/b.go\n@@ -10,0 +11 @@\n+z\n+++ /dev/null\n"
	d, err := guard.Parse(status, numstat, patch)
	if err != nil {
		t.Fatal(err)
	}
	want := guard.Diff{Files: []guard.File{
		{Path: "a.go", Status: 'M', Removed: 1, Added: []guard.Line{{N: 3, Text: "x := 1"}, {N: 4, Text: "y := 2"}}},
		{Path: "new/b.go", OldPath: "old/b.go", Status: 'R', Removed: 3, Added: []guard.Line{{N: 11, Text: "z"}}},
		{Path: "c_test.go", Status: 'D', Removed: 9},
	}}
	if diff := cmp.Diff(want, d); diff != "" {
		t.Errorf("Parse (-want +got):\n%s", diff)
	}
}

func TestParseRefusesAnUnknownStatusLine(t *testing.T) {
	if _, err := guard.Parse("garbage\n", "", ""); err == nil {
		t.Fatal("Parse accepted a status line with no path")
	}
}

func rules(fs []guard.Finding) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return out
}

func TestAnApprovedChangeMayTouchTestsButNeverBypass(t *testing.T) {
	d := guard.Diff{Files: []guard.File{
		{Path: "x_test.go", Status: 'M', Added: []guard.Line{{N: 4, Text: "\tt.Skip()"}}},
		{Path: "go.mod", Status: 'M', Added: []guard.Line{{N: 3, Text: "go 1.26.0"}}},
	}}
	if diff := cmp.Diff([]string{"bypass-skip", "protected-test", "protected-gomod"}, rules(guard.Check(d, false))); diff != "" {
		t.Errorf("unapproved (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"bypass-skip"}, rules(guard.Check(d, true))); diff != "" {
		t.Errorf("approved (-want +got):\n%s", diff)
	}
}

func TestBuildTagsOfRealTargetsPass(t *testing.T) {
	for _, tag := range []string{"//go:build windows", "//go:build !windows && (amd64 || arm64)", "//go:build go1.25 && unix", "// +build linux"} {
		d := guard.Diff{Files: []guard.File{{Path: "a.go", Status: 'A', Added: []guard.Line{{N: 1, Text: tag}}}}}
		if fs := guard.Check(d, false); len(fs) != 0 {
			t.Errorf("%s: %v", tag, rules(fs))
		}
	}
}

func TestADeletedFileIsNeverABypass(t *testing.T) {
	d := guard.Diff{Files: []guard.File{{Path: "_old/a.go", Status: 'D', Removed: 3}}}
	if fs := guard.Check(d, true); len(fs) != 0 {
		t.Errorf("Check = %v", rules(fs))
	}
}

func TestTheModuleIsTheFirstInternalSegment(t *testing.T) {
	added := []guard.Line{{N: 1, Text: "package x"}}
	d := guard.Diff{Files: []guard.File{
		{Path: "internal/notes/internal/big/a.go", Status: 'A', Added: added},
		{Path: "internal/notes/b.go", Status: 'A', Added: added},
	}}
	if fs := guard.Check(d, false); len(fs) != 0 {
		t.Errorf("one module counted as several: %v", fs)
	}
}
