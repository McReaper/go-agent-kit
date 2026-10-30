package list_test

import (
	"context"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/list"
	"github.com/google/go-cmp/cmp"
)

func TestListFiltersByTagBeforeTheLimit(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "n1", Title: "A", Tags: []domain.Tag{"home"}},
		{ID: "n2", Title: "B", Tags: []domain.Tag{"work"}},
		{ID: "n3", Title: "C", Tags: []domain.Tag{"work", "urgent"}},
		{ID: "n4", Title: "D", Tags: []domain.Tag{"work"}},
	}}
	out, err := list.New(store).Run(context.Background(), list.Args{Tag: "work", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	want := list.Out{Notes: []list.Item{
		{ID: "n2", Title: "B", Tags: []domain.Tag{"work"}},
		{ID: "n3", Title: "C", Tags: []domain.Tag{"work", "urgent"}},
	}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestListWithATagNoNoteCarriesIsEmpty(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "n1", Title: "A"}}}
	out, err := list.New(store).Run(context.Background(), list.Args{Tag: "work"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 0 {
		t.Errorf("Notes = %+v, want none", out.Notes)
	}
}

func TestListRefusesAnInvalidTagFilter(t *testing.T) {
	if _, err := list.New(&memStore{}).Run(context.Background(), list.Args{Tag: "Work"}); err == nil {
		t.Error("tag filter \"Work\" accepted")
	}
}
