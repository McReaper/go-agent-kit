package remove_test

import (
	"context"
	"errors"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/remove"
)

type memStore struct {
	deleted []domain.NoteID
	err     error
}

func (s *memStore) Delete(_ context.Context, id domain.NoteID) error {
	if s.err != nil {
		return s.err
	}
	s.deleted = append(s.deleted, id)
	return nil
}

func TestDeleteRemovesTheNoteAndReturnsItsID(t *testing.T) {
	store := &memStore{}
	out, err := remove.New(store).Run(context.Background(), remove.Args{ID: "buy-milk-20261030t220509"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "buy-milk-20261030t220509" {
		t.Errorf("ID = %q", out.ID)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "buy-milk-20261030t220509" {
		t.Errorf("deleted %v", store.deleted)
	}
}

func TestDeleteRefusesAnIDThatIsNotAFileNameWithoutDeleting(t *testing.T) {
	for _, id := range []string{"", "../notes", "Buy Milk"} {
		store := &memStore{}
		_, err := remove.New(store).Run(context.Background(), remove.Args{ID: id})
		if !errors.Is(err, domain.ErrEmptyID) && !errors.Is(err, domain.ErrBadID) {
			t.Errorf("Run(%q) err = %v, want an id error", id, err)
		}
		if len(store.deleted) != 0 {
			t.Errorf("Run(%q) deleted %v", id, store.deleted)
		}
	}
}

func TestDeleteReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk gone")
	_, err := remove.New(&memStore{err: boom}).Run(context.Background(), remove.Args{ID: "x-1"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
