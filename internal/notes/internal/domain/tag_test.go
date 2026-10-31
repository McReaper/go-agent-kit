package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
)

func TestParseTagLowercasesAndTrims(t *testing.T) {
	got, err := domain.ParseTag("  Go ")
	if err != nil || got != "go" {
		t.Fatalf("ParseTag = %q, %v; want %q, nil", got, err, "go")
	}
}

func TestParseTagRefusesAnEmptyTag(t *testing.T) {
	if _, err := domain.ParseTag("   "); !errors.Is(err, domain.ErrEmptyTag) {
		t.Fatalf("err = %v, want ErrEmptyTag", err)
	}
}

func TestParseTagRefusesATagOverTheLimit(t *testing.T) {
	if _, err := domain.ParseTag(strings.Repeat("x", domain.MaxTagRunes+1)); err == nil {
		t.Fatal("a tag over the limit was accepted")
	}
	if _, err := domain.ParseTag(strings.Repeat("x", domain.MaxTagRunes)); err != nil {
		t.Fatalf("a tag of exactly the limit was refused: %v", err)
	}
}

func TestParseTagsKeepsEachTagOnceIgnoringCase(t *testing.T) {
	got, err := domain.ParseTags([]string{"Go", "rust", " GO "})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "go" || got[1] != "rust" {
		t.Fatalf("ParseTags = %v, want [go rust]", got)
	}
}

func TestParseTagsOfNoneIsNil(t *testing.T) {
	got, err := domain.ParseTags(nil)
	if err != nil || got != nil {
		t.Fatalf("ParseTags(nil) = %#v, %v; want nil, nil", got, err)
	}
}

func TestParseTagsRefusesOneInvalidTag(t *testing.T) {
	if _, err := domain.ParseTags([]string{"ok", ""}); !errors.Is(err, domain.ErrEmptyTag) {
		t.Fatalf("err = %v, want ErrEmptyTag", err)
	}
}

func TestANoteHasTheTagsItWasFiledUnder(t *testing.T) {
	n := domain.Note{Tags: []domain.Tag{"go"}}
	if !n.HasTag("go") || n.HasTag("rust") {
		t.Errorf("HasTag: go = %v, rust = %v; want true, false", n.HasTag("go"), n.HasTag("rust"))
	}
}
