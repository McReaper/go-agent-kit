package remove_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/remove"
)

type memStore struct {
	deleted []domain.NoteID
	err     error
}

func (s *memStore) Delete(_ context.Context, id domain.NoteID) error {
	s.deleted = append(s.deleted, id)
	return s.err
}

func TestDeleteRemovesTheNoteWithTheGivenID(t *testing.T) {
	store := &memStore{}
	out, err := remove.New(store).Run(context.Background(), remove.Args{ID: "buy-milk-20261030t220509"})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "buy-milk-20261030t220509" {
		t.Errorf("deleted %v, want the one note buy-milk-20261030t220509", store.deleted)
	}
	if out.ID != "buy-milk-20261030t220509" {
		t.Errorf("returned id %q, want the id removed", out.ID)
	}
}

func TestDeleteRefusesAMalformedID(t *testing.T) {
	for _, id := range []string{"", "Buy Milk", "../escape", "buy-milk-"} {
		store := &memStore{}
		if _, err := remove.New(store).Run(context.Background(), remove.Args{ID: id}); err == nil {
			t.Errorf("id %q accepted", id)
		}
		if len(store.deleted) != 0 {
			t.Errorf("id %q reached the store: %v", id, store.deleted)
		}
	}
}

func TestDeleteSaysSoWhenNoNoteHasTheID(t *testing.T) {
	store := &memStore{err: domain.ErrNoNote}
	_, err := remove.New(store).Run(context.Background(), remove.Args{ID: "buy-milk-20261030t220509"})
	if !errors.Is(err, domain.ErrNoNote) {
		t.Fatalf("err = %v, want it to wrap ErrNoNote", err)
	}
	if !strings.Contains(err.Error(), "no note has that id") {
		t.Errorf("err = %q, want it to say no note has that id", err)
	}
}

func TestDeleteReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk gone")
	_, err := remove.New(&memStore{err: boom}).Run(context.Background(), remove.Args{ID: "buy-milk-20261030t220509"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
