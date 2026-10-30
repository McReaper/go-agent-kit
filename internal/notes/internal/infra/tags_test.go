package infra_test

import (
	"context"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"github.com/google/go-cmp/cmp"
)

func TestFileStoreKeepsTheTags(t *testing.T) {
	s := open(t)
	ctx := context.Background()
	n := note("tagged", 4)
	n.Tags = []domain.Tag{"work", "2026"}
	if err := s.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	got, err := s.Newest(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Tag{"work", "2026"}, got[0].Tags); diff != "" {
		t.Errorf("tags (-want +got):\n%s", diff)
	}
}
