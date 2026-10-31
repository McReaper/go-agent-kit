package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/google/go-cmp/cmp"
)

func TestParseTagLowercasesAndTrims(t *testing.T) {
	got, err := domain.ParseTag("  Go ")
	if err != nil || got != "go" {
		t.Fatalf("ParseTag = %q, %v; want %q, nil", got, err, "go")
	}
}

func TestParseTagRefusesEmptyAndLongTags(t *testing.T) {
	if _, err := domain.ParseTag("   "); !errors.Is(err, domain.ErrEmptyTag) {
		t.Errorf("empty: err = %v, want ErrEmptyTag", err)
	}
	if _, err := domain.ParseTag(strings.Repeat("é", domain.MaxTagRunes+1)); err == nil {
		t.Error("a tag over the limit was accepted")
	}
	if _, err := domain.ParseTag(strings.Repeat("é", domain.MaxTagRunes)); err != nil {
		t.Errorf("a tag of exactly the limit: err = %v", err)
	}
}

func TestParseTagsDropsRepeatsIgnoringCase(t *testing.T) {
	got, err := domain.ParseTags([]string{"Go", "go", "Rust", " RUST ", "tools"})
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Tag{"go", "rust", "tools"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseTags (-want +got):\n%s", diff)
	}
}

func TestParseTagsIsNilWithoutTags(t *testing.T) {
	for _, in := range [][]string{nil, {}} {
		got, err := domain.ParseTags(in)
		if err != nil || got != nil {
			t.Errorf("ParseTags(%v) = %v, %v; want nil, nil", in, got, err)
		}
	}
}

func TestParseTagsRefusesAnEmptyTagWhole(t *testing.T) {
	got, err := domain.ParseTags([]string{"ok", ""})
	if !errors.Is(err, domain.ErrEmptyTag) || got != nil {
		t.Errorf("ParseTags = %v, %v; want nil and ErrEmptyTag", got, err)
	}
}

func TestHasTagMatchesOnlyCarriedTags(t *testing.T) {
	n := domain.Note{Tags: []domain.Tag{"go"}}
	if !n.HasTag("go") || n.HasTag("rust") {
		t.Errorf("HasTag: go=%v rust=%v", n.HasTag("go"), n.HasTag("rust"))
	}
}
