// Package mutesting runs go-mutesting and reads the report it writes.
package mutesting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Mutant is one mutation: where it starts, the mutator, the first line it
// wrote (empty when it only removed code), and whether a test killed it.
type Mutant struct {
	File    string
	Line    int
	Mutator string
	Wrote   string
	Killed  bool
}

// Run mutates the packages in dir, given as ./dir patterns, and returns the
// mutants. go-mutesting writes each mutant over the source while the tests
// run: dir must be a copy.
//
// WHY: go-mutesting exits 1 when a mutant lived; any other failure means it
// did not run, which must not pass as a perfect score.
func Run(ctx context.Context, dir string, pkgs []string) ([]Mutant, error) {
	out, err := run.Command(ctx, dir, nil, append([]string{"go-mutesting"}, pkgs...)...)
	if err != nil {
		return nil, err
	}
	if out.Code > 1 {
		return nil, fmt.Errorf("go-mutesting: exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	return Read(dir)
}

type entry struct {
	Mutator struct {
		Name string `json:"mutatorName"`
		File string `json:"originalFilePath"`
		Line int    `json:"originalStartLine"`
	} `json:"mutator"`
	Diff string `json:"diff"`
}

type report struct {
	Escaped  []entry `json:"escaped"`
	Killed   []entry `json:"killed"`
	TimedOut []entry `json:"timeouted"`
}

// Read reads the report.json go-mutesting wrote in dir: a timed out mutant
// counts as killed.
func Read(dir string) (ms []Mutant, err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dir, err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	data, err := root.ReadFile("report.json")
	if err != nil {
		return nil, fmt.Errorf("read go-mutesting's report: %w", err)
	}
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("decode go-mutesting's report: %w", err)
	}
	for _, e := range r.Escaped {
		ms = append(ms, e.mutant(false))
	}
	for _, e := range slices.Concat(r.Killed, r.TimedOut) {
		ms = append(ms, e.mutant(true))
	}
	return ms, nil
}

func (e entry) mutant(killed bool) Mutant {
	m := e.Mutator
	return Mutant{File: filepath.ToSlash(m.File), Line: m.Line, Mutator: m.Name, Wrote: wrote(e.Diff), Killed: killed}
}

func wrote(diff string) string {
	for l := range strings.Lines(diff) {
		if strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++") {
			return strings.TrimSpace(l[1:])
		}
	}
	return ""
}
