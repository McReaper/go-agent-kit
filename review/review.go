// Package review writes a change's review sheet: the few lines its owner
// reads instead of the diff. First what the change does to the spec, then
// what else needs their approval; or that nothing does, and the change
// merges by itself once its checks pass.
package review

import (
	"context"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/spec"
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
	Spec     []spec.Change
	Approval []Item // the harness, a dependency, the go line, a library's API, the size
}

// Build reads the change from base to the working tree of the git
// repository at dir, the module's root, and writes its sheet.
func Build(ctx context.Context, dir, base string) (Sheet, error) {
	_, fs, err := guard.Run(ctx, dir, base)
	if err != nil {
		return Sheet{}, err
	}
	var s Sheet
	if s.Spec, err = guard.SpecChanges(ctx, dir, base); err != nil {
		return Sheet{}, err
	}
	for _, f := range fs {
		if f.NeedsApproval() && f.Rule != "protected-spec" {
			s.Approval = append(s.Approval, Item{File: f.File, Line: f.Line, Text: f.Subject()})
		}
	}
	return s, nil
}

// Empty reports whether the change needs nothing from the owner.
func (s Sheet) Empty() bool {
	return len(s.Spec)+len(s.Approval) == 0
}

// Labels are the pull request labels the kit sets, each with whether this
// sheet earns it: the caller adds the ones that are true and removes the
// others.
func (s Sheet) Labels() map[string]bool {
	return map[string]bool{"needs-you": !s.Empty()}
}
