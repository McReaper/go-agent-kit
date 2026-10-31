package remove_test

import (
	"context"
	"errors"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/remove"
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

func TestDeleteRemovesTheNoteByID(t *testing.T) {
	store := &memStore{}
	out, err := remove.New(store).Run(context.Background(), remove.Args{ID: "buy-milk-20261030t220509"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "buy-milk-20261030t220509" || len(store.deleted) != 1 || store.deleted[0] != "buy-milk-20261030t220509" {
		t.Errorf("Run = %+v, deleted %v", out, store.deleted)
	}
}

func TestDeleteRefusesAnIDNoteCanHaveWithoutDeleting(t *testing.T) {
	store := &memStore{}
	_, err := remove.New(store).Run(context.Background(), remove.Args{ID: "../escape"})
	if !errors.Is(err, domain.ErrBadID) {
		t.Fatalf("err = %v, want ErrBadID", err)
	}
	if len(store.deleted) != 0 {
		t.Errorf("deleted %v", store.deleted)
	}
}

func TestDeleteSaysWhenNoNoteHasTheID(t *testing.T) {
	store := &memStore{err: domain.ErrNoSuchNote}
	_, err := remove.New(store).Run(context.Background(), remove.Args{ID: "gone-20261030t220509"})
	if !errors.Is(err, domain.ErrNoSuchNote) {
		t.Fatalf("err = %v, want ErrNoSuchNote", err)
	}
}
