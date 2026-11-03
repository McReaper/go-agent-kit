// Package spec reads an app's spec, the Gherkin scenarios under features/
// that its owner approves and reads instead of the code, and checks it two
// ways: what a change does to it (Changes), and whether each scenario can
// fail (Prove).
package spec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v42"
	messages "github.com/cucumber/messages/go/v34"

	"github.com/McReaper/go-agent-kit/internal/platform/git"
)

// Dir is where an app keeps its spec, from the module's root.
const Dir = "features"

// WIP tags a scenario the owner approved that isn't built yet: the spec's
// test leaves it out, and the change that builds it removes the tag.
const WIP = "@wip"

// Kind is what a step does: sets the scene, acts, or checks.
type Kind int

// The kinds of step, after their Gherkin keywords; And and But take the
// kind of the step before them.
const (
	Given Kind = iota
	When
	Then
)

// Step is one step of a scenario, its background's included.
type Step struct {
	Line      int
	Kind      Kind
	Text      string
	DocString *string
	Table     [][]string
}

// Scenario is one scenario of the spec; each example row of an outline is
// one.
type Scenario struct {
	File  string // from the module's root, like features/notes.feature
	Line  int
	Name  string
	WIP   bool
	Tags  []string // its tags but @wip
	Steps []Step
}

// Same reports whether s and o say the same thing: same steps, same tags,
// wherever they sit in their files. @wip is not part of what they say.
func (s Scenario) Same(o Scenario) bool {
	return slices.Equal(s.Tags, o.Tags) && slices.EqualFunc(s.Steps, o.Steps, sameStep)
}

func sameStep(a, b Step) bool {
	return a.Kind == b.Kind && a.Text == b.Text &&
		(a.DocString == nil) == (b.DocString == nil) && (a.DocString == nil || *a.DocString == *b.DocString) &&
		slices.EqualFunc(a.Table, b.Table, slices.Equal)
}

// Read reads the spec in the working tree of the module at dir. A module
// without a features directory has no spec.
func Read(dir string) (out []Scenario, err error) {
	root, err := os.OpenRoot(path.Join(dir, Dir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the spec: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	err = fs.WalkDir(root.FS(), ".", func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || path.Ext(p) != ".feature" {
			return err
		}
		data, err := root.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		s, err := Parse(path.Join(Dir, p), data)
		out = append(out, s...)
		return err
	})
	return out, err
}

// ReadAt reads the spec as it is at ref in the git repository at dir.
func ReadAt(ctx context.Context, dir, ref string) ([]Scenario, error) {
	if err := git.CheckRef(ref); err != nil {
		return nil, err
	}
	list, err := git.Run(ctx, dir, "ls-tree", "-r", "-z", "--name-only", ref, "--", Dir+"/")
	if err != nil {
		return nil, err
	}
	var out []Scenario
	for p := range strings.SplitSeq(strings.TrimSuffix(list, "\x00"), "\x00") {
		if path.Ext(p) != ".feature" {
			continue
		}
		text, err := git.Run(ctx, dir, "show", ref+":./"+p)
		if err != nil {
			return nil, err
		}
		s, err := Parse(p, []byte(text))
		if err != nil {
			return nil, err
		}
		out = append(out, s...)
	}
	return out, nil
}

// Parse reads the scenarios of one feature file, file being its path from
// the module's root.
func Parse(file string, text []byte) ([]Scenario, error) {
	ids := &messages.Incrementing{}
	doc, err := gherkin.ParseGherkinDocument(bytes.NewReader(text), ids.NewId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if doc.Feature == nil {
		return nil, nil
	}
	lines := stepLines(doc.Feature)
	var out []Scenario
	for _, p := range gherkin.Pickles(*doc, file, ids.NewId) {
		out = append(out, scenarioOf(file, p, lines))
	}
	return out, nil
}

func scenarioOf(file string, p *messages.Pickle, lines map[string]int) Scenario {
	s := Scenario{File: file, Name: p.Name}
	if p.Location != nil {
		s.Line = int(p.Location.Line)
	}
	for _, t := range p.Tags {
		if t.Name == WIP {
			s.WIP = true
		} else {
			s.Tags = append(s.Tags, t.Name)
		}
	}
	for _, ps := range p.Steps {
		s.Steps = append(s.Steps, stepOf(ps, lines))
	}
	return s
}

func stepOf(ps *messages.PickleStep, lines map[string]int) Step {
	st := Step{Text: ps.Text, Kind: kindOf(ps.Type)}
	if len(ps.AstNodeIds) > 0 {
		st.Line = lines[ps.AstNodeIds[0]]
	}
	if a := ps.Argument; a != nil && a.DocString != nil {
		st.DocString = &a.DocString.Content
	}
	if a := ps.Argument; a != nil && a.DataTable != nil {
		for _, r := range a.DataTable.Rows {
			st.Table = append(st.Table, cellsOf(r))
		}
	}
	return st
}

func cellsOf(r *messages.PickleTableRow) []string {
	row := make([]string, 0, len(r.Cells))
	for _, c := range r.Cells {
		row = append(row, c.Value)
	}
	return row
}

func kindOf(t messages.PickleStepType) Kind {
	switch t {
	case messages.PickleStepType_ACTION:
		return When
	case messages.PickleStepType_OUTCOME:
		return Then
	case messages.PickleStepType_CONTEXT, messages.PickleStepType_UNKNOWN:
	}
	return Given
}

type lineIndex map[string]int

func stepLines(f *messages.Feature) lineIndex {
	lines := lineIndex{}
	for _, c := range f.Children {
		lines.add(c.Background, c.Scenario)
		lines.addRule(c.Rule)
	}
	return lines
}

func (l lineIndex) addRule(r *messages.Rule) {
	if r == nil {
		return
	}
	for _, c := range r.Children {
		l.add(c.Background, c.Scenario)
	}
}

func (l lineIndex) add(bg *messages.Background, sc *messages.Scenario) {
	var steps []*messages.Step
	switch {
	case bg != nil:
		steps = bg.Steps
	case sc != nil:
		steps = sc.Steps
	}
	for _, s := range steps {
		l[s.Id] = int(s.Location.Line)
	}
}
