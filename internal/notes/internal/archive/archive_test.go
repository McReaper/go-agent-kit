package archive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/archive"
	"example.com/notes/internal/notes/internal/domain"
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

const id = "buy-milk-20261030t220509"

func withNote(t *testing.T) *memStore {
	t.Helper()
	n := domain.New("Buy milk", "two litres", time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC))
	return &memStore{notes: map[domain.NoteID]domain.Note{n.ID: n}}
}

func TestArchiveHidesTheNoteAndKeepsEverythingElse(t *testing.T) {
	store := withNote(t)
	out, err := archive.New(store).ArchiveOp().Run(context.Background(), archive.Args{ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if out != (archive.Out{ID: id, Title: "Buy milk"}) {
		t.Errorf("Out = %+v", out)
	}
	want := store.notes[id]
	want.Archived = true
	if diff := cmp.Diff([]domain.Note{want}, store.replaced); diff != "" {
		t.Errorf("replaced (-want +got):\n%s", diff)
	}
}

func TestUnarchiveShowsTheNoteAgain(t *testing.T) {
	store := withNote(t)
	a := archive.New(store)
	if _, err := a.ArchiveOp().Run(context.Background(), archive.Args{ID: id}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.UnarchiveOp().Run(context.Background(), archive.Args{ID: id}); err != nil {
		t.Fatal(err)
	}
	if got := store.notes[id]; got.Archived || got.Title != "Buy milk" || got.Body != "two litres" {
		t.Errorf("after unarchive: %+v", got)
	}
}

func TestArchivingTwiceChangesNothingMore(t *testing.T) {
	store := withNote(t)
	a := archive.New(store)
	for range 2 {
		if _, err := a.ArchiveOp().Run(context.Background(), archive.Args{ID: id}); err != nil {
			t.Fatalf("archive: %v", err)
		}
	}
	if got := store.notes[id]; !got.Archived {
		t.Errorf("note not archived: %+v", got)
	}
}

func TestArchiveRefusesWithoutAnIDAndDoesNotTouchTheStore(t *testing.T) {
	store := withNote(t)
	if _, err := archive.New(store).ArchiveOp().Run(context.Background(), archive.Args{}); err == nil {
		t.Fatal("archive accepted an empty id")
	}
	if _, err := archive.New(store).UnarchiveOp().Run(context.Background(), archive.Args{}); err == nil {
		t.Fatal("unarchive accepted an empty id")
	}
	if len(store.replaced) != 0 {
		t.Errorf("replaced %d notes", len(store.replaced))
	}
}

func TestArchiveReportsAnUnknownIDAsNotFound(t *testing.T) {
	_, err := archive.New(withNote(t)).ArchiveOp().Run(context.Background(), archive.Args{ID: "nope"})
	if !errors.Is(err, domain.ErrNoteNotFound) {
		t.Fatalf("err = %v, want ErrNoteNotFound", err)
	}
}

func TestArchiveReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk full")
	store := withNote(t)
	store.err = boom
	_, err := archive.New(store).UnarchiveOp().Run(context.Background(), archive.Args{ID: id})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
