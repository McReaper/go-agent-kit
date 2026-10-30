package infra_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/infra"
	"github.com/google/go-cmp/cmp"
)

func open(t *testing.T) *infra.FileStore {
	t.Helper()
	s, err := infra.OpenFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

func note(title string, minute int) domain.Note {
	return domain.New(domain.Title(title), "body of "+title, time.Date(2026, 10, 30, 22, minute, 0, 0, time.UTC))
}

func TestFileStoreReturnsTheNewestFirstUpToTheLimit(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	for i, title := range []string{"old", "newest", "middle"} {
		if err := s.Save(ctx, note(title, []int{1, 3, 2}[i])); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Newest(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Note{note("newest", 3), note("middle", 2)}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Newest (-want +got):\n%s", diff)
	}
}

func TestFileStoreDeletesOnlyTheNamedNote(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	for _, minute := range []int{1, 2} {
		if err := s.Save(ctx, note("note", minute)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Save(ctx, note("other", 3)); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, note("note", 2).ID); err != nil {
		t.Fatal(err)
	}
	got, err := s.Newest(ctx, 10)
	want := []domain.Note{note("other", 3), note("note", 1)}
	if err != nil || len(got) != len(want) {
		t.Fatalf("Newest = %+v, %v", got, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Newest after Delete (-want +got):\n%s", diff)
	}
}

func TestFileStoreRefusesToDeleteAMissingNote(t *testing.T) {
	s := open(t)
	err := s.Delete(context.Background(), note("ghost", 1).ID)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestFileStoreNeverOverwritesANote(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	if err := s.Save(ctx, note("same", 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctx, note("same", 1)); err == nil {
		t.Fatal("the second save of the same id succeeded")
	}
	got, err := s.Newest(ctx, 10)
	if err != nil || len(got) != 1 || got[0].Body != "body of same" {
		t.Fatalf("Newest = %+v, %v", got, err)
	}
}
