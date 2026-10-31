package search_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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

func TestSearchFindsTheNotesWhoseTitleContainsTheText(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "buy-milk-20261030t220509", Title: "Buy milk"},
		{ID: "read-book-20261030t220509", Title: "Read book", Body: "about dairy"},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk"})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{{ID: "buy-milk-20261030t220509", Title: "Buy milk"}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchFindsTheNotesWhoseBodyContainsTheText(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "read-book-20261030t220509", Title: "Read book", Body: "about dairy"},
		{ID: "buy-milk-20261030t220509", Title: "Buy milk"},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "dairy"})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{{ID: "read-book-20261030t220509", Title: "Read book", Body: "about dairy"}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchIgnoresCaseInTheTitleAndTheBody(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "Buy MILK"},
		{ID: "b-2", Title: "Read", Body: "Some Dairy"},
	}}
	for _, text := range []string{"milk", "MILK", "dAiRy"} {
		out, err := search.New(store).Run(context.Background(), search.Args{Text: text})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Notes) != 1 {
			t.Errorf("search %q found %d notes, want 1", text, len(out.Notes))
		}
	}
}

func TestSearchReturnsTheNewestFirst(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "milk-new-20261030t220509", Title: "Milk, new"},
		{ID: "milk-old-20261029t220509", Title: "Milk, old"},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 2 || out.Notes[0].ID != "milk-new-20261030t220509" || out.Notes[1].ID != "milk-old-20261029t220509" {
		t.Errorf("Run = %+v, want the newer note first", out.Notes)
	}
}

func TestSearchLooksAtEveryNote(t *testing.T) {
	var notes []domain.Note
	for i := range 25 {
		notes = append(notes, domain.Note{ID: domain.NoteID(fmt.Sprintf("n-%d", i)), Title: "Filler"})
	}
	notes = append(notes, domain.Note{ID: "old-milk", Title: "Old milk"})
	out, err := search.New(&memStore{notes: notes}).Run(context.Background(), search.Args{Text: "milk"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 1 || out.Notes[0].ID != "old-milk" {
		t.Errorf("Run = %+v, want the note past the first 25 found", out.Notes)
	}
}

func TestSearchReturnsNoNotesWhenNoneMatches(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "a-1", Title: "A"}}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "zebra"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Notes == nil || len(out.Notes) != 0 {
		t.Errorf("Run = %+v, want an empty list", out.Notes)
	}
}

func TestSearchRefusesAnEmptyText(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "a-1", Title: "A"}}}
	for _, text := range []string{"", "   "} {
		_, err := search.New(store).Run(context.Background(), search.Args{Text: text})
		if !errors.Is(err, search.ErrEmptyText) {
			t.Errorf("text %q: err = %v, want ErrEmptyText", text, err)
		}
	}
}

func TestSearchKeepsTheSpacesInsideTheText(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "a-1", Title: "Buy milk"}, {ID: "b-2", Title: "Milk Buy"}}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "buy milk"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 1 || out.Notes[0].ID != "a-1" {
		t.Errorf("Run = %+v, want only a-1: the text is matched whole", out.Notes)
	}
}

func TestSearchReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk gone")
	_, err := search.New(&memStore{err: boom}).Run(context.Background(), search.Args{Text: "milk"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestSearchReturnsTheTagsOfEachFoundNote(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "Buy milk", Tags: []domain.Tag{"errands", "todo"}},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk"})
	if err != nil {
		t.Fatal(err)
	}
	want := search.Out{Notes: []search.Item{{ID: "a-1", Title: "Buy milk", Tags: []string{"errands", "todo"}}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestSearchReturnsTwentyMatchesByDefault(t *testing.T) {
	var notes []domain.Note
	for i := range 25 {
		notes = append(notes, domain.Note{ID: domain.NoteID(fmt.Sprintf("milk-%d", i)), Title: "Milk"})
	}
	out, err := search.New(&memStore{notes: notes}).Run(context.Background(), search.Args{Text: "milk"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 20 {
		t.Errorf("Run returned %d notes, want 20 by default", len(out.Notes))
	}
}

func TestSearchCountsTheLimitAmongTheMatches(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "Filler"},
		{ID: "b-2", Title: "Milk"},
		{ID: "c-3", Title: "Filler"},
		{ID: "d-4", Title: "Milk"},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk", Limit: new(1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 1 || out.Notes[0].ID != "b-2" {
		t.Errorf("Run = %+v, want the newest match b-2", out.Notes)
	}
}

func TestSearchRefusesALimitOutOfRange(t *testing.T) {
	for _, n := range []int{0, -1, 101} {
		if _, err := search.New(&memStore{}).Run(context.Background(), search.Args{Text: "milk", Limit: new(n)}); err == nil {
			t.Errorf("limit %d accepted", n)
		}
	}
}

func TestSearchAcceptsTheLimitsAtTheirEnds(t *testing.T) {
	var notes []domain.Note
	for i := range 101 {
		notes = append(notes, domain.Note{ID: domain.NoteID(fmt.Sprintf("milk-%d", i)), Title: "Milk"})
	}
	for _, limit := range []int{1, 100} {
		out, err := search.New(&memStore{notes: notes}).Run(context.Background(), search.Args{Text: "milk", Limit: new(limit)})
		if err != nil {
			t.Fatalf("limit %d refused: %v", limit, err)
		}
		if len(out.Notes) != limit {
			t.Errorf("limit %d returned %d notes", limit, len(out.Notes))
		}
	}
}
