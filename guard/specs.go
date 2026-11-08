package guard

import (
	"context"
	"fmt"

	"github.com/McReaper/rumblestrip-go/spec"
)

// SpecChanges returns what the change from base to the working tree of the
// repository at dir does to the spec.
//
// WHY: the spec is read at base's tip, not at the merge base: the question
// is what merging changes on base, and a branch stacked on a spec branch
// that has merged since holds that spec already.
func SpecChanges(ctx context.Context, dir, base string) ([]spec.Change, error) {
	before, err := spec.ReadAt(ctx, dir, base)
	if err != nil {
		return nil, err
	}
	after, err := spec.Read(dir)
	if err != nil {
		return nil, err
	}
	return spec.Changes(before, after), nil
}

func specChanges(ctx context.Context, dir, base string) ([]Finding, error) {
	cs, err := SpecChanges(ctx, dir, base)
	if err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(cs))
	for _, c := range cs {
		out = append(out, Finding{"protected-spec", c.File, c.Line, fmt.Sprintf(
			"T2: the spec changes, scenario %q %s: the owner approves the spec; building a scenario, removing its @wip tag, is free", c.Name, c.Kind)})
	}
	return out, nil
}
