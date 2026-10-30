package edit_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/edit"
	"github.com/google/go-cmp/cmp"
)

type memStore struct {
	notes    map[domain.NoteID]domain.Note
	replaced []domain.Note
	err      error
}

func (s *memStore) Get(_ context.Context, id domain.NoteID) (domain.Note, error) {
	n, ok := s.notes[id]
	if !ok {
		return domain.Note{}, fmt.Errorf("%w: %s", domain.ErrNoteNotFound, id)
	}
	return n, nil
}

func (s *memStore) Replace(_ context.Context, n domain.Note) error {
	if s.err != nil {
		return s.err
	}
	s.replaced = append(s.replaced, n)
	s.notes[n.ID] = n
	return nil
}

var created = time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC)

func withNote(t *testing.T) *memStore {
	t.Helper()
	n := domain.New("Buy milk", "two litres", created)
	return &memStore{notes: map[domain.NoteID]domain.Note{n.ID: n}}
}

const id = "buy-milk-20261030t220509"

func TestEditChangesTheTitleAndKeepsIDAndDate(t *testing.T) {
	store := withNote(t)
	out, err := edit.New(store).Run(context.Background(), edit.Args{ID: id, Title: " Buy oat milk "})
	if err != nil {
		t.Fatal(err)
	}
	if out != (edit.Out{ID: id, Title: "Buy oat milk"}) {
		t.Errorf("Out = %+v", out)
	}
	want := domain.Note{ID: id, Title: "Buy oat milk", Body: "two litres", Created: created}
	if diff := cmp.Diff([]domain.Note{want}, store.replaced); diff != "" {
		t.Errorf("replaced (-want +got):\n%s", diff)
	}
}

func TestEditChangesTheBodyAndKeepsTheTitle(t *testing.T) {
	store := withNote(t)
	if _, err := edit.New(store).Run(context.Background(), edit.Args{ID: id, Body: "three litres"}); err != nil {
		t.Fatal(err)
	}
	got := store.replaced[0]
	if got.Title != "Buy milk" || got.Body != "three litres" {
		t.Errorf("replaced %+v", got)
	}
}

func TestEditClearBodyEmptiesTheText(t *testing.T) {
	store := withNote(t)
	if _, err := edit.New(store).Run(context.Background(), edit.Args{ID: id, ClearBody: true}); err != nil {
		t.Fatal(err)
	}
	if got := store.replaced[0]; got.Body != "" || got.Title != "Buy milk" {
		t.Errorf("replaced %+v", got)
	}
}

func TestEditRefusesInvalidArgumentsWithoutTouchingTheNote(t *testing.T) {
	cases := []struct {
		name string
		args edit.Args
	}{
		{"no id", edit.Args{Title: "x"}},
		{"nothing to change", edit.Args{ID: id}},
		{"body and clear_body", edit.Args{ID: id, Body: "x", ClearBody: true}},
		{"an empty title after trimming", edit.Args{ID: id, Title: "   "}},
		{"a title with nothing to name it by", edit.Args{ID: id, Title: "?!"}},
		{"a title over the limit", edit.Args{ID: id, Title: strings.Repeat("a", 81)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := withNote(t)
			if _, err := edit.New(store).Run(context.Background(), c.args); err == nil {
				t.Fatal("accepted")
			}
			if len(store.replaced) != 0 {
				t.Errorf("replaced %d notes", len(store.replaced))
			}
		})
	}
}

func TestEditReportsAnUnknownIDAsNotFound(t *testing.T) {
	store := withNote(t)
	_, err := edit.New(store).Run(context.Background(), edit.Args{ID: "nope", Title: "x"})
	if !errors.Is(err, domain.ErrNoteNotFound) {
		t.Fatalf("err = %v, want ErrNoteNotFound", err)
	}
}

func TestEditReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk full")
	store := withNote(t)
	store.err = boom
	_, err := edit.New(store).Run(context.Background(), edit.Args{ID: id, Title: "x"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
