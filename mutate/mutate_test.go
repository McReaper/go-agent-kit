package mutate_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/mutate"
)

func TestChangedLinesAreTheProductionGoLinesAChangeAdds(t *testing.T) {
	d := guard.Diff{Files: []guard.File{
		{Path: "a/a.go", Status: 'M', Added: []guard.Line{{N: 3}, {N: 7}}},
		{Path: "a/a_test.go", Status: 'M', Added: []guard.Line{{N: 1}}},
		{Path: "a/testdata/x.go", Status: 'A', Added: []guard.Line{{N: 1}}},
		{Path: "README.md", Status: 'M', Added: []guard.Line{{N: 1}}},
		{Path: "b/gone.go", Status: 'D'},
		{Path: "b/b.go", Status: 'A', Added: []guard.Line{{N: 1}}},
	}}
	got := mutate.Changed(d)
	want := mutate.Lines{"a/a.go": {3, 7}, "b/b.go": {1}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("changed lines (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"./a", "./b"}, mutate.Packages(got)); diff != "" {
		t.Errorf("packages (-want +got):\n%s", diff)
	}
}

func TestOnlyMutantsStartingOnAChangedLineCount(t *testing.T) {
	ms := []mutate.Mutant{
		{File: "a/a.go", Line: 3, Killed: true},
		{File: "a/a.go", Line: 4},
		{File: "a/a.go", Line: 7},
		{File: "b/b.go", Line: 3},
	}
	r := mutate.Score(ms, mutate.Lines{"a/a.go": {3, 7}})
	want := mutate.Report{Mutants: []mutate.Mutant{ms[0], ms[2]}, Score: 0.5}
	if diff := cmp.Diff(want, r); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
}

func TestAChangeWithNothingToMutateScoresFull(t *testing.T) {
	ms := []mutate.Mutant{{File: "a/a.go", Line: 4}}
	if r := mutate.Score(ms, mutate.Lines{"a/a.go": {3}}); r.Score != 1 || len(r.Mutants) != 0 {
		t.Errorf("got %+v, want no mutant and score 1", r)
	}
}
