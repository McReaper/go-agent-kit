package infra_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/infra"
	"github.com/google/go-cmp/cmp"
)

func TestFileStoreRoundTripsTags(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	n := note("tagged", 4)
	n.Tags = []domain.Tag{"go", "tools"}
	if err := s.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctx, note("plain", 5)); err != nil {
		t.Fatal(err)
	}
	got, err := s.Newest(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Note{note("plain", 5), n}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Newest (-want +got):\n%s", diff)
	}
}

func TestFileStoreRefusesARecordWithABadTag(t *testing.T) {
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
	bad := `{"id":"fine-1","title":"Fine","body":"","tags":["  "],"created":"2026-10-30T22:00:00Z"}`
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Newest(context.Background(), 10); err == nil {
		t.Fatal("Newest accepted a record with an empty tag")
	}
}
