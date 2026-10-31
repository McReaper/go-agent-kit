package mutate_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/McReaper/go-agent-kit/mutate"
)

func TestAMutantThatOnlyDropsAnErrorReturnIsNotCounted(t *testing.T) {
	ms := []mutate.Mutant{
		{File: "a/a.go", Line: 3, Killed: true},
		{File: "a/a.go", Line: 4, DropsErrorReturn: true},
		{File: "a/a.go", Line: 5},
		{File: "b/b.go", Line: 4, DropsErrorReturn: true},
	}
	got := mutate.Score(ms, mutate.Lines{"a/a.go": {3, 4, 5}})
	want := mutate.Report{Mutants: []mutate.Mutant{ms[0], ms[2]}, Score: 0.5, ErrorReturns: 1}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
	if r := mutate.Score(ms[1:2], mutate.Lines{"a/a.go": {4}}); r.Score != 1 || r.ErrorReturns != 1 {
		t.Errorf("only an error return changed: %+v, want score 1", r)
	}
}
