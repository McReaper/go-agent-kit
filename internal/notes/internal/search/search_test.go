package search_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/search"
	"github.com/google/go-cmp/cmp"
)

type memStore struct {
	notes []domain.Note
	err   error
}

func (s *memStore) Newest(_ context.Context, limit int) ([]domain.Note, error) {
	return s.notes[:min(limit, len(s.notes))], s.err
}

var at = time.Date(2026, 10, 30, 22, 0, 0, 0, time.UTC)

func note(id, title, body string, ago int) domain.Note {
	return domain.Note{ID: domain.NoteID(id), Title: domain.Title(title), Body: body, Created: at.Add(-time.Duration(ago) * time.Minute)}
}

func TestSearchMatchesTitleOrBodyIgnoringCase(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		note("a-1", "Buy MILK", "", 0),
		note("b-2", "Groceries", "two litres of milk", 1),
		note("c-3", "Unrelated", "bread", 2),
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "Milk"})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{{ID: "a-1", Title: "Buy MILK"}, {ID: "b-2", Title: "Groceries"}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchKeepsNewestFirstAndTags(t *testing.T) {
	older := note("old-1", "Ship it", "release", 5)
	older.Tags = []domain.Tag{"go", "release"}
	store := &memStore{notes: []domain.Note{note("new-2", "Release day", "", 0), older}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "release"})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{
		{ID: "new-2", Title: "Release day"},
		{ID: "old-1", Title: "Ship it", Tags: []string{"go", "release"}},
	}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchWithNoMatchReturnsAnEmptyList(t *testing.T) {
	store := &memStore{notes: []domain.Note{note("a-1", "A", "b", 0)}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "zzz"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Notes == nil || len(out.Notes) != 0 {
		t.Errorf("Notes = %#v, want an empty non-nil list", out.Notes)
	}
}

func TestSearchRefusesAnEmptyText(t *testing.T) {
	if _, err := search.New(&memStore{}).Run(context.Background(), search.Args{}); err == nil {
		t.Error("empty text accepted")
	}
}

func TestSearchReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk gone")
	_, err := search.New(&memStore{err: boom}).Run(context.Background(), search.Args{Text: "x"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
