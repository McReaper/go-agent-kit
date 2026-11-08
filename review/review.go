// Package review writes a change's review sheet: the few lines its owner
// reads instead of the diff. First where the change stands, then what it
// does to the spec, then what else needs their approval; or that nothing
// does, and the change merges by itself once its checks pass.
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

// State is where the owner's approval stands.
type State string

// The owner's approval: nothing needs it, it is awaited, or given.
const (
	NothingToApprove State = "none"
	Waiting          State = "waiting"
	Approved         State = "approved"
)

// Sheet is a change's review sheet.
type Sheet struct {
	Spec     []spec.Change
	Approval []Item // the harness, a dependency, the go line, a library's API, the size
	State    State
	theirs   []guard.Finding
}

// Build reads the change from base to the working tree of the git
// repository at dir, the module's root, and writes its sheet. Its State is
// Waiting while anything is the owner's: Approve settles it.
func Build(ctx context.Context, dir, base string) (Sheet, error) {
	_, fs, err := guard.Run(ctx, dir, base)
	if err != nil {
		return Sheet{}, err
	}
	s := Sheet{State: NothingToApprove}
	if s.Spec, err = guard.SpecChanges(ctx, dir, base); err != nil {
		return Sheet{}, err
	}
	for _, f := range fs {
		if !f.NeedsApproval() {
			continue
		}
		s.theirs = append(s.theirs, f)
		s.State = Waiting
		if f.Rule != "protected-spec" {
			s.Approval = append(s.Approval, Item{File: f.File, Line: f.Line, Text: f.Subject()})
		}
	}
	return s, nil
}

// Approve returns the sheet with its State Approved once a, what a code
// owner approved, covers everything that is theirs (see guard.Covered).
func (s Sheet) Approve(ctx context.Context, a guard.Approval) (Sheet, error) {
	if s.State != Waiting {
		return s, nil
	}
	covered, err := guard.Covered(ctx, s.theirs, a)
	if err != nil {
		return Sheet{}, err
	}
	if len(covered) == len(s.theirs) {
		s.State = Approved
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
	return map[string]bool{"needs-you": s.State == Waiting}
}
