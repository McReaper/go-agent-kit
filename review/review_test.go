package review_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/review"
	"github.com/McReaper/go-agent-kit/spec"
)

func TestASheetWithNothingForTheOwnerSaysTheChangeMergesByItself(t *testing.T) {
	s := review.Sheet{State: review.NothingToApprove}
	want := review.Marker + "\n" + review.StatusStart + "\n⏳ The gates are running. Nothing for you: the kit merges it once they pass.\n" + review.StatusEnd + "\n" +
		"\n<sub>[go-agent-kit](https://github.com/McReaper/go-agent-kit) v1.2.3</sub>\n"
	if got := s.Markdown("v1.2.3"); got != want {
		t.Errorf("Markdown() = %q, want %q", got, want)
	}
	if s.Labels()["needs-you"] {
		t.Error("an empty sheet earns needs-you")
	}
}

func TestASheetListsTheSpecFirstThenWhatElseNeedsApproval(t *testing.T) {
	s := review.Sheet{
		Spec:     []spec.Change{{Kind: spec.Disabled, File: "features/notes.feature", Line: 7, Name: "A note is listed"}},
		Approval: []review.Item{{File: "go.mod", Line: 9, Text: "a new dependency, example.com/x"}, {Text: "this change touches 401 production lines, the limit is 400"}},
		State:    review.Waiting,
	}
	got := s.Markdown("")
	for _, want := range []string{
		"**Waiting for you:** read what is below, then approve the pull request.\n",
		"- A note is listed, disabled: tagged `@wip`, it stops running (`features/notes.feature:7`)\n",
		"- a new dependency, example.com/x (`go.mod:9`)\n",
		"- this change touches 401 production lines, the limit is 400 (`the change`)\n",
		"<sub>[go-agent-kit](https://github.com/McReaper/go-agent-kit)</sub>\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the sheet lacks %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "The spec") > strings.Index(got, "Also yours") {
		t.Errorf("the spec comes after the rest:\n%s", got)
	}
	if !s.Labels()["needs-you"] {
		t.Error("a sheet waiting for the owner lacks needs-you")
	}
}

// WHY: the label tells the owner which pull requests wait for them; once
// they approved, it goes.
func TestTheStatusAndTheLabelFollowTheOwnersApproval(t *testing.T) {
	for _, c := range []struct {
		state    review.State
		says     string
		needsYou bool
	}{
		{review.NothingToApprove, "Nothing for you: the kit merges it once they pass.", false},
		{review.Waiting, "**Waiting for you:**", true},
		{review.Approved, "You approved what is yours: the kit merges it once they pass.", false},
	} {
		s := review.Sheet{State: c.state}
		if !strings.Contains(s.Status(), c.says) || s.Labels()["needs-you"] != c.needsYou {
			t.Errorf("%s: status %q, needs-you %v", c.state, s.Status(), s.Labels()["needs-you"])
		}
	}
}

// WHY: GitHub notifies the account an @name outside code names: a sheet
// saying a scenario got its @wip tag would ping someone called wip.
func TestTheSheetMentionsNoOne(t *testing.T) {
	var changes []spec.Change
	for _, k := range []spec.ChangeKind{spec.Added, spec.Changed, spec.Removed, spec.Disabled} {
		changes = append(changes, spec.Change{Kind: k, File: "features/notes.feature", Line: 7, Name: "A note is listed"})
	}
	got := review.Sheet{Spec: changes, State: review.Waiting}.Markdown("v1.0.0")
	outside := regexp.MustCompile("`[^`]*`").ReplaceAllString(got, "")
	if strings.Contains(outside, "@") {
		t.Errorf("the sheet has an @ outside code:\n%s", got)
	}
}
