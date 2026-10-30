package infra_test

import (
	"context"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"github.com/google/go-cmp/cmp"
)

func archived(n domain.Note) domain.Note {
	n.Archived = true
	return n
}

func TestFileStoreKeepsTheArchivedFlag(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	n := archived(note("archived", 6))
	if err := s.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(n, got); diff != "" {
		t.Errorf("Get (-want +got):\n%s", diff)
	}
}

func TestFileStoreRecentLeavesOutArchivedNotesAndCountsOnlyTheRest(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	for _, n := range []domain.Note{note("old", 1), note("middle", 2), archived(note("hidden", 3)), archived(note("newest", 4))} {
		if err := s.Save(ctx, n); err != nil {
			t.Fatal(err)
		}
	}

	live, err := s.Recent(ctx, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Note{note("middle", 2), note("old", 1)}, live); diff != "" {
		t.Errorf("Recent without archived (-want +got):\n%s", diff)
	}

	all, err := s.Recent(ctx, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Note{archived(note("newest", 4)), archived(note("hidden", 3))}, all); diff != "" {
		t.Errorf("Recent with archived (-want +got):\n%s", diff)
	}
}
