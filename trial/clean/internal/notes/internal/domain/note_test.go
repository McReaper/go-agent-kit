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
	t.Run("refuses an empty title", func(t *testing.T) {
		if _, err := domain.ParseTitle("   "); !errors.Is(err, domain.ErrEmptyTitle) {
			t.Fatalf("err = %v, want ErrEmptyTitle", err)
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
