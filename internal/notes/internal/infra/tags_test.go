package infra_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/infra"
	"github.com/google/go-cmp/cmp"
)

func TestFileStoreKeepsTheTagsOfANote(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	n := note("tagged", 4)
	n.Tags = []domain.Tag{"go", "todo"}
	if err := s.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	got, err := s.Newest(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Note{n}, got); diff != "" {
		t.Errorf("Newest (-want +got):\n%s", diff)
	}
}

func TestFileStoreRefusesAStoredTagItCannotParse(t *testing.T) {
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
	bad := `{"id":"fine-20261030t220000","title":"Fine","body":"","created":"2026-10-30T22:00:00Z","tags":["  "]}`
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Newest(context.Background(), 10); !errors.Is(err, domain.ErrEmptyTag) {
		t.Fatalf("Newest err = %v, want ErrEmptyTag", err)
	}
}
