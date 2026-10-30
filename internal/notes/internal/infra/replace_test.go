package infra_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/infra"
	"github.com/google/go-cmp/cmp"
)

func TestFileStoreGetReturnsTheNoteByID(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	n := note("groceries", 4)
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

func TestFileStoreGetReportsAnUnknownIDAsNotFound(t *testing.T) {
	_, err := open(t).Get(context.Background(), "nope-20261030t220104")
	if !errors.Is(err, domain.ErrNoteNotFound) {
		t.Fatalf("err = %v, want ErrNoteNotFound", err)
	}
}

func TestFileStoreReplaceRewritesOnlyThatNote(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	kept := note("kept", 1)
	changed := note("changed", 2)
	for _, n := range []domain.Note{kept, changed} {
		if err := s.Save(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	changed.Title = "renamed"
	changed.Body = "new text"
	if err := s.Replace(ctx, changed); err != nil {
		t.Fatal(err)
	}
	got, err := s.Newest(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Note{changed, kept}, got); diff != "" {
		t.Errorf("Newest after Replace (-want +got):\n%s", diff)
	}
}

func TestFileStoreReplaceRefusesAMissingNote(t *testing.T) {
	dir := t.TempDir()
	s, err := infra.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	ghost := note("ghost", 3)
	if err := s.Replace(context.Background(), ghost); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want fs.ErrNotExist", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, string(ghost.ID)+".json")); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("Replace created the note: %v", statErr)
	}
}

func TestFileStoreReplaceLeavesNoTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	s, err := infra.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx := context.Background()
	n := note("tidy", 5)
	if err := s.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	if err := s.Replace(ctx, n); err != nil {
		t.Fatal(err)
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Errorf("temporary files left behind: %v", matches)
	}
}
