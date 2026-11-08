package spec_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/McReaper/rumblestrip-go/spec"
)

const notes = `Feature: Notes

  Background:
    Given an empty notebook

  Scenario: A note is listed
    When I add the note "Groceries"
    Then the list shows 1 note "Groceries"

  @wip @slow
  Scenario Outline: A long title is refused
    When I add a note of <n> characters
    Then it is refused with "<message>"

    Examples:
      | n  | message          |
      | 81 | the limit is 80  |
      | 99 | the limit is 80  |
`

func parse(t *testing.T, text string) []spec.Scenario {
	t.Helper()
	s, err := spec.Parse("features/notes.feature", []byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEachRowOfAnOutlineIsAScenarioWithItsValuesAndTheBackground(t *testing.T) {
	s := parse(t, notes)
	if len(s) != 3 {
		t.Fatalf("got %d scenarios, want 3", len(s))
	}
	got := s[2]
	if got.Name != "A long title is refused" || !got.WIP || !slices.Equal(got.Tags, []string{"@slow"}) || got.Line != 18 {
		t.Errorf("the second row: %+v", got)
	}
	want := []spec.Step{
		{Line: 4, Kind: spec.Given, Text: "an empty notebook"},
		{Line: 12, Kind: spec.When, Text: "I add a note of 99 characters"},
		{Line: 13, Kind: spec.Then, Text: `it is refused with "the limit is 80"`},
	}
	if diff := cmp.Diff(want, got.Steps); diff != "" {
		t.Errorf("steps (-want +got):\n%s", diff)
	}
}

func TestChangesNamesWhatTheOwnerApprovesAndLetsABuiltScenarioGo(t *testing.T) {
	before := parse(t, notes)
	after := parse(t, `Feature: Notes

  Background:
    Given an empty notebook

  @wip
  Scenario: A note is listed
    When I add the note "Groceries"
    Then the list shows 1 note "Groceries"

  @slow
  Scenario Outline: A long title is refused
    When I add a note of <n> characters
    Then it is refused with "<message>"

    Examples:
      | n  | message          |
      | 81 | the limit is 80  |
      | 99 | the limit is 90  |

  Scenario: A note is deleted
    When I delete the note "Groceries"
    Then the list shows 0 notes
`)
	got := spec.Changes(before, after)
	want := []spec.Change{
		{Kind: spec.Disabled, File: "features/notes.feature", Line: 7, Name: "A note is listed"},
		{Kind: spec.Changed, File: "features/notes.feature", Line: 19, Name: "A long title is refused"},
		{Kind: spec.Added, File: "features/notes.feature", Line: 21, Name: "A note is deleted"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("changes (-want +got):\n%s", diff)
	}
	if got := spec.Changes(after, after[:3]); len(got) != 1 || got[0].Kind != spec.Removed || got[0].Name != "A note is deleted" {
		t.Errorf("a scenario taken out: %+v", got)
	}
}

func TestAScenarioWrittenBackAsGherkinReadsTheSame(t *testing.T) {
	for _, s := range parse(t, notes+`
  Scenario: With a table and a doc string
    Then the notes are
      | title     |
      | Groceries |
    And the journal says
      """
      added "Groceries"
      """
`) {
		back := parse(t, s.Gherkin())
		if len(back) != 1 || !back[0].Same(s) || back[0].WIP {
			t.Errorf("%q written back reads %+v", s.Name, back)
		}
	}
}
