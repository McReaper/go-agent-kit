package review_test

import (
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/review"
	"github.com/McReaper/go-agent-kit/spec"
)

func TestASheetWithNothingForTheOwnerSaysTheChangeMergesByItself(t *testing.T) {
	var s review.Sheet
	if got := s.Markdown(); got != review.Marker+"\nNothing for you: this pull request merges by itself once its checks pass.\n" {
		t.Errorf("Markdown() = %q", got)
	}
	if s.Labels()["needs-you"] {
		t.Error("an empty sheet earns needs-you")
	}
}

func TestASheetListsTheSpecFirstThenWhatElseNeedsApproval(t *testing.T) {
	s := review.Sheet{
		Spec:     []spec.Change{{Kind: spec.Disabled, File: "features/notes.feature", Line: 7, Name: "A note is listed"}},
		Approval: []review.Item{{File: "go.mod", Line: 9, Text: "a new dependency, example.com/x"}, {Text: "this change touches 401 production lines, the limit is 400"}},
	}
	got := s.Markdown()
	for _, want := range []string{
		"- A note is listed, disabled: tagged @wip, it stops running (`features/notes.feature:7`)\n",
		"- a new dependency, example.com/x (`go.mod:9`)\n",
		"- this change touches 401 production lines, the limit is 400 (`the change`)\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the sheet lacks %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "The spec") > strings.Index(got, "Also yours") {
		t.Errorf("the spec comes after the rest:\n%s", got)
	}
	if !s.Labels()["needs-you"] {
		t.Error("a sheet with items lacks needs-you")
	}
}
