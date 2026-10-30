package add_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/add"
	"example.com/notes/internal/notes/internal/domain"
)

type memStore struct {
	saved []domain.Note
	err   error
}

func (s *memStore) Save(_ context.Context, n domain.Note) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, n)
	return nil
}

func fixedClock() time.Time {
	return time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC)
}

func TestAddSavesTheNoteAndReturnsItsID(t *testing.T) {
	store := &memStore{}
	out, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: " Buy milk ", Body: "two litres"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "buy-milk-20261030t220509" {
		t.Errorf("ID = %q", out.ID)
	}
	if len(store.saved) != 1 || store.saved[0].Title != "Buy milk" || store.saved[0].Body != "two litres" {
		t.Errorf("saved %+v", store.saved)
	}
}

func TestAddRefusesAnInvalidTitleWithoutSaving(t *testing.T) {
	store := &memStore{}
	_, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: "  "})
	if !errors.Is(err, domain.ErrEmptyTitle) {
		t.Fatalf("err = %v, want ErrEmptyTitle", err)
	}
	if len(store.saved) != 0 {
		t.Errorf("saved %d notes", len(store.saved))
	}
}

func TestAddReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk full")
	_, err := add.New(&memStore{err: boom}, fixedClock).Run(context.Background(), add.Args{Title: "x"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
