package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"example.com/notes/internal/notes/internal/domain"
	"pgregory.net/rapid"
)

func TestParseTitle(t *testing.T) {
	t.Run("trims the spaces around it", func(t *testing.T) {
		got, err := domain.ParseTitle("  Groceries  ")
		if err != nil || got != "Groceries" {
			t.Fatalf("ParseTitle = %q, %v; want %q, nil", got, err, "Groceries")
		}
	})
	t.Run("refuses a title with nothing to name the note by", func(t *testing.T) {
		if _, err := domain.ParseTitle("?!"); !errors.Is(err, domain.ErrNoSlugInTitle) {
			t.Fatalf("err = %v, want ErrNoSlugInTitle", err)
		}
	})
	t.Run("refuses a title over the limit, counted in characters", func(t *testing.T) {
		long := strings.Repeat("é", domain.MaxTitleRunes+1)
		_, err := domain.ParseTitle(long)
		if err == nil || !strings.Contains(err.Error(), "81 characters") {
			t.Fatalf("err = %v, want the count of characters", err)
		}
	})
}

func TestParseTitleAcceptsEveryTrimmedTitleWithinTheLimit(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.StringMatching(`[a-zA-Z0-9][a-zA-Z0-9 ]{0,78}[a-zA-Z0-9]`).Draw(t, "title")
		got, err := domain.ParseTitle(s)
		if err != nil || string(got) != s {
			t.Fatalf("ParseTitle(%q) = %q, %v", s, got, err)
		}
		if utf8.RuneCountInString(string(got)) > domain.MaxTitleRunes {
			t.Fatalf("accepted %d characters", utf8.RuneCountInString(string(got)))
		}
	})
}

func TestNewNamesTheNoteByItsTitleAndUTCTime(t *testing.T) {
	at := time.Date(2026, 10, 30, 23, 5, 9, 0, time.FixedZone("CET", 3600))
	n := domain.New("Buy milk", "two litres", at)
	if n.ID != "buy-milk-20261030t220509" {
		t.Errorf("ID = %q", n.ID)
	}
	if !n.Created.Equal(at) || n.Created.Location() != time.UTC {
		t.Errorf("Created = %v, want %v in UTC", n.Created, at)
	}
}

func TestParseNoteID(t *testing.T) {
	for _, ok := range []string{"buy-milk-20261030t220509", "a", "a1-b2"} {
		if got, err := domain.ParseNoteID(ok); err != nil || string(got) != ok {
			t.Errorf("ParseNoteID(%q) = %q, %v", ok, got, err)
		}
	}
	for _, bad := range []string{"", "-a", "a-", "a--b", "A", "../x", "a b", "a/b"} {
		if _, err := domain.ParseNoteID(bad); !errors.Is(err, domain.ErrBadID) {
			t.Errorf("ParseNoteID(%q) err = %v, want ErrBadID", bad, err)
		}
	}
}

func TestEveryNewNoteHasAnIDThatParses(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		title := rapid.StringMatching(`[a-zA-Z0-9][a-zA-Z0-9 ]{0,40}`).Draw(t, "title")
		n := domain.New(domain.Title(title), "", time.Unix(rapid.Int64Range(0, 4e9).Draw(t, "at"), 0))
		if _, err := domain.ParseNoteID(string(n.ID)); err != nil {
			t.Fatalf("New(%q) made id %q: %v", title, n.ID, err)
		}
	})
}

func FuzzParseTitle(f *testing.F) {
	for _, seed := range []string{"Groceries", "", "  x  ", "?!", strings.Repeat("a", 81)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, err := domain.ParseTitle(s)
		if err != nil {
			return
		}
		if string(got) != strings.TrimSpace(string(got)) || utf8.RuneCountInString(string(got)) > domain.MaxTitleRunes {
			t.Fatalf("ParseTitle(%q) accepted %q", s, got)
		}
	})
}
