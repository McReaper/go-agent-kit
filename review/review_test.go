package review_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/review"
)

func TestATestNameReadsAsASentence(t *testing.T) {
	for name, want := range map[string]string{
		"TestAddReportsTheStoreFailure":     "Add reports the store failure",
		"TestKitMutateFailsAChange":         "Kit mutate fails a change",
		"TestJSONFileIsRead":                "JSON file is read",
		"TestV2APIAnswers":                  "V2 API answers",
		"Test_add_note":                     "Add note",
		"TestParseTitle/trims_the_spaces":   "Parse title: trims the spaces",
		"TestParseTitle/refuses an empty":   "Parse title: refuses an empty",
		"FuzzSlugIsASCII":                   "Fuzz: Slug is ASCII",
		"ExampleMake":                       "Example: Make",
		"Test":                              "",
		"TestParseNoteID":                   "Parse note ID",
		"TestAddsTheNoteIn3Steps":           "Adds the note in3 steps",
		"TestEveryNewNoteHasAnIDThatParses": "Every new note has an ID that parses",
	} {
		if got := review.Sentence(name); got != want {
			t.Errorf("Sentence(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestAnEmptySheetSaysThereIsNothingToReview(t *testing.T) {
	var s review.Sheet
	if got := s.Markdown(); got != review.Marker+"\n## Review sheet\n\nNothing beyond the code itself: no protected change, no test scenario, no contract, dependency, shared code or rule touched.\n" {
		t.Errorf("Markdown() = %q", got)
	}
	for name, on := range s.Labels() {
		if on {
			t.Errorf("label %s on an empty sheet", name)
		}
	}
}
