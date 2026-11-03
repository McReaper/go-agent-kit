package spec

import "strconv"

// Change is what a change does to one scenario of the spec, one the owner
// approves. A scenario that loses its @wip tag and nothing else is built,
// not changed: it is no change.
type Change struct {
	Kind ChangeKind
	File string
	Line int // in the spec after the change, or before it for a removal
	Name string
}

// ChangeKind is what happened to a scenario.
type ChangeKind string

// The changes the owner approves.
const (
	Added    ChangeKind = "added"
	Changed  ChangeKind = "changed"
	Removed  ChangeKind = "removed"
	Disabled ChangeKind = "disabled" // tagged @wip again: it stops running
)

// Changes returns what going from before to after does to the spec, one
// change per scenario: an outline's rows count as their scenario once.
func Changes(before, after []Scenario) []Change {
	was, wasKeys := index(before)
	is, isKeys := index(after)
	var out changes
	for i, s := range after {
		out.add(changeOf(was, isKeys[i], s))
	}
	for i, s := range before {
		if _, ok := is[wasKeys[i]]; !ok {
			out.add(&Change{Removed, s.File, s.Line, s.Name})
		}
	}
	return out
}

func changeOf(was map[string]Scenario, key string, s Scenario) *Change {
	old, ok := was[key]
	switch {
	case !ok:
		return &Change{Added, s.File, s.Line, s.Name}
	case !old.Same(s):
		return &Change{Changed, s.File, s.Line, s.Name}
	case s.WIP && !old.WIP:
		return &Change{Disabled, s.File, s.Line, s.Name}
	}
	return nil
}

type changes []Change

func (cs *changes) add(c *Change) {
	if c == nil {
		return
	}
	for _, o := range *cs {
		if o.Kind == c.Kind && o.File == c.File && o.Name == c.Name {
			return
		}
	}
	*cs = append(*cs, *c)
}

// WHY: an outline's rows share their scenario's name; the n-th row with a
// name is that name's n-th scenario, before and after alike.
func index(spec []Scenario) (map[string]Scenario, []string) {
	byKey := map[string]Scenario{}
	keys := make([]string, len(spec))
	seen := map[string]int{}
	for i, s := range spec {
		name := s.File + "\x00" + s.Name
		keys[i] = name + "\x00" + strconv.Itoa(seen[name])
		seen[name]++
		byKey[keys[i]] = s
	}
	return byKey, keys
}
