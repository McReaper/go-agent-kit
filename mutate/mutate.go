// Package mutate scores the tests of a change: it mutates the lines the
// change touched and counts how many mutants a test kills (criterion M6).
// go-mutesting writes each mutant over the source while the tests run, so
// it only ever runs on a copy of the module.
package mutate

import (
	"context"
	"errors"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/internal/platform/mutesting"
)

// Mutant is one mutation of a changed line.
type Mutant = mutesting.Mutant

// Report is the mutants of the changed lines and their score: the share a
// test killed, 1 when no changed line could be mutated. ErrorReturns counts
// the mutants left out because they only drop an error's return.
type Report struct {
	Mutants      []Mutant
	Score        float64
	ErrorReturns int
}

// Run mutates the lines changed since base in the module at dir, the root
// of its git repository. A mutant counts when it starts on a changed line.
func Run(ctx context.Context, dir, base string) (r Report, err error) {
	d, err := guard.Read(ctx, dir, base)
	if err != nil {
		return Report{}, err
	}
	changed := Changed(d)
	if len(changed) == 0 {
		return Report{Score: 1}, nil
	}
	work, err := copyModule(ctx, dir)
	if err != nil {
		return Report{}, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(work)) }()
	mutants, err := mutesting.Run(ctx, work, Packages(changed))
	if err != nil {
		return Report{}, err
	}
	return Score(mutants, changed), nil
}

// Lines are the lines a change adds to each production Go file, by path.
type Lines map[string][]int

// Changed returns the lines d adds to production Go files: tests,
// testdata and deleted files left out.
func Changed(d guard.Diff) Lines {
	out := Lines{}
	for _, f := range d.Files {
		if f.Status == 'D' || path.Ext(f.Path) != ".go" || strings.HasSuffix(f.Path, "_test.go") || strings.Contains("/"+f.Path, "/testdata/") {
			continue
		}
		for _, l := range f.Added {
			out[f.Path] = append(out[f.Path], l.N)
		}
	}
	return out
}

// Packages returns the directories of the files in changed, as ./dir
// patterns, sorted.
func Packages(changed Lines) []string {
	var dirs []string
	for file := range changed {
		if d := "./" + path.Dir(file); !slices.Contains(dirs, d) {
			dirs = append(dirs, d)
		}
	}
	slices.Sort(dirs)
	return dirs
}

// Score keeps the mutants that start on a changed line and scores them.
//
// WHY: a mutant that only removes `if err != nil { return …err… }` lives
// unless a test makes the call fail, which for I/O rarely pays; and the
// compiler (an unused err) and errcheck already refuse a dropped error.
// Counted, they take the template's infra from 91% to 55% (ADR 4).
func Score(ms []Mutant, changed Lines) Report {
	var r Report
	killed := 0
	for _, m := range ms {
		if !slices.Contains(changed[m.File], m.Line) {
			continue
		}
		if m.DropsErrorReturn {
			r.ErrorReturns++
			continue
		}
		r.Mutants = append(r.Mutants, m)
		if m.Killed {
			killed++
		}
	}
	r.Score = 1
	if len(r.Mutants) > 0 {
		r.Score = float64(killed) / float64(len(r.Mutants))
	}
	return r
}
