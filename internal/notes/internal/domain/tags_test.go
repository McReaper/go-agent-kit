package domain_test

import (
	"errors"
	"strings"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"github.com/google/go-cmp/cmp"
)

func TestParseTagAcceptsLowercaseLettersAndDigits(t *testing.T) {
	for _, s := range []string{"a", "work", "2026", "q3report"} {
		got, err := domain.ParseTag(s)
		if err != nil || string(got) != s {
			t.Errorf("ParseTag(%q) = %q, %v", s, got, err)
		}
	}
}

func TestParseTagRefusesEverythingElse(t *testing.T) {
	if _, err := domain.ParseTag(""); !errors.Is(err, domain.ErrEmptyTag) {
		t.Errorf("empty tag: err = %v, want ErrEmptyTag", err)
	}
	cases := map[string]string{
		"Work":     "W",
		" work":    " ",
		"new-york": "-",
		"a_b":      "_",
		"café":     "é",
	}
	for s, bad := range cases {
		if _, err := domain.ParseTag(s); err == nil || !strings.Contains(err.Error(), bad) {
			t.Errorf("ParseTag(%q) err = %v, want a refusal naming %q", s, err, bad)
		}
	}
}

func TestParseTagsKeepsTheOrderAndAllowsNone(t *testing.T) {
	got, err := domain.ParseTags([]string{"b", "a"})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Tag{"b", "a"}, got); diff != "" {
		t.Errorf("ParseTags (-want +got):\n%s", diff)
	}
	if got, err := domain.ParseTags(nil); err != nil || len(got) != 0 {
		t.Errorf("ParseTags(nil) = %v, %v", got, err)
	}
}

func TestParseTagsCapsTheCount(t *testing.T) {
	if _, err := domain.ParseTags([]string{"a", "b", "c", "d", "e"}); err != nil {
		t.Errorf("five tags refused: %v", err)
	}
	_, err := domain.ParseTags([]string{"a", "b", "c", "d", "e", "f"})
	if err == nil || !strings.Contains(err.Error(), "at most 5") {
		t.Errorf("six tags: err = %v, want the limit of 5", err)
	}
}

func TestParseTagsRefusesADuplicate(t *testing.T) {
	_, err := domain.ParseTags([]string{"work", "home", "work"})
	if err == nil || !strings.Contains(err.Error(), `"work" is given twice`) {
		t.Errorf("err = %v, want the duplicate named", err)
	}
}

func TestHasTagFindsOnlyTheNotesTags(t *testing.T) {
	n := domain.Note{Tags: []domain.Tag{"work"}}
	if !n.HasTag("work") || n.HasTag("home") {
		t.Errorf("HasTag on %v gave the wrong answers", n.Tags)
	}
}
