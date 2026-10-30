package search_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/search"
	"github.com/google/go-cmp/cmp"
)

type memStore struct {
	notes []domain.Note
	err   error
}

func (s *memStore) All(_ context.Context) ([]domain.Note, error) {
	return s.notes, s.err
}

func at(minute int) time.Time {
	return time.Date(2026, 10, 30, 22, minute, 0, 0, time.UTC)
}

func TestSearchFindsTheNotesWhoseTitleOrTextHasTheText(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "Buy milk", Body: "not here"},
		{ID: "b-2", Title: "Groceries", Body: "a MILK carton"},
		{ID: "c-3", Title: "Bread", Body: "rye"},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk"})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{{ID: "a-1", Title: "Buy milk"}, {ID: "b-2", Title: "Groceries"}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchKeepsTheStoreOrderAndStopsAtTheLimit(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "new-3", Title: "milk 3", Created: at(3)},
		{ID: "mid-2", Title: "milk 2", Created: at(2)},
		{ID: "old-1", Title: "milk 1", Created: at(1)},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{{ID: "new-3", Title: "milk 3"}, {ID: "mid-2", Title: "milk 2"}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchReturnsAnEmptyListWhenNothingMatches(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "a-1", Title: "Buy milk"}}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "bread"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Notes == nil || len(out.Notes) != 0 {
		t.Errorf("Notes = %#v, want an empty non-nil list", out.Notes)
	}
}

func TestSearchRefusesAnEmptyText(t *testing.T) {
	for _, text := range []string{"", "   "} {
		if _, err := search.New(&memStore{}).Run(context.Background(), search.Args{Text: text}); err == nil {
			t.Errorf("text %q accepted", text)
		}
	}
}

func TestSearchRefusesALimitOutOfRange(t *testing.T) {
	for _, limit := range []int{-1, 101} {
		if _, err := search.New(&memStore{}).Run(context.Background(), search.Args{Text: "x", Limit: limit}); err == nil {
			t.Errorf("limit %d accepted", limit)
		}
	}
}

func TestSearchReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk gone")
	_, err := search.New(&memStore{err: boom}).Run(context.Background(), search.Args{Text: "x"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
