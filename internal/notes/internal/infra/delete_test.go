package infra_test

import (
	"context"
	"errors"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
)

func TestFileStoreDeletesOnlyTheNoteAsked(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	keep, gone := note("keep", 1), note("gone", 2)
	for _, n := range []domain.Note{keep, gone} {
		if err := s.Save(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Delete(ctx, gone.ID); err != nil {
		t.Fatal(err)
	}
	got, err := s.Newest(ctx, 10)
	if err != nil || len(got) != 1 || got[0].ID != keep.ID {
		t.Fatalf("Newest after delete = %+v, %v; want only %s", got, err, keep.ID)
	}
}

func TestFileStoreSaysWhenNoNoteHasTheID(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	if err := s.Save(ctx, note("here", 1)); err != nil {
		t.Fatal(err)
	}
	err := s.Delete(ctx, "not-here-20261030t220101")
	if !errors.Is(err, domain.ErrNoSuchNote) {
		t.Fatalf("Delete of a missing id: err = %v, want ErrNoSuchNote", err)
	}
	got, err := s.Newest(ctx, 10)
	if err != nil || len(got) != 1 {
		t.Fatalf("Newest = %+v, %v; the existing note must stay", got, err)
	}
}
