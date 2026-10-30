package list_test

import (
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/list"
	"github.com/google/go-cmp/cmp"
)

func TestSearchFindsAPartOfATitle(t *testing.T) {
	notes := []domain.Note{{ID: "buy-milk-1", Title: "Buy milk"}, {ID: "call-mum-2", Title: "Call mum"}}
	want := []list.Item{{ID: "buy-milk-1", Title: "Buy milk"}}
	if diff := cmp.Diff(want, list.Search(notes, "mil")); diff != "" {
		t.Errorf("Search (-want +got):\n%s", diff)
	}
}
