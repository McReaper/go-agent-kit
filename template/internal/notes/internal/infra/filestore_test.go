package infra_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/infra"
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

func TestFileStoreRefusesARecordWithABadID(t *testing.T) {
	dir := t.TempDir()
	s, err := infra.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	bad := `{"id":"../escape","title":"Fine","body":"","created":"2026-10-30T22:00:00Z"}`
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Newest(context.Background(), 10); !errors.Is(err, domain.ErrBadID) {
		t.Fatalf("Newest err = %v, want ErrBadID", err)
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
