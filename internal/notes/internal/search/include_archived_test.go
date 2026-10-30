package search_test

import (
	"context"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/search"
	"github.com/google/go-cmp/cmp"
)

func TestSearchLeavesOutArchivedNotesUnlessAskedFor(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "milk now", Archived: true},
		{ID: "b-2", Title: "milk later"},
		{ID: "c-3", Title: "milk old", Archived: true},
	}}
	cases := []struct {
		include bool
		want    []search.Item
	}{
		{false, []search.Item{{ID: "b-2", Title: "milk later"}}},
		{true, []search.Item{{ID: "a-1", Title: "milk now"}, {ID: "b-2", Title: "milk later"}, {ID: "c-3", Title: "milk old"}}},
	}
	for _, c := range cases {
		out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk", IncludeArchived: c.include})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(search.Out{Notes: c.want}, out); diff != "" {
			t.Errorf("include_archived=%v (-want +got):\n%s", c.include, diff)
		}
	}
}

func TestSearchLimitCountsOnlyTheNotesShown(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "milk hidden", Archived: true},
		{ID: "b-2", Title: "milk shown"},
		{ID: "c-3", Title: "milk also shown"},
	}}
	out, err := search.New(store).Run(context.Background(), search.Args{Text: "milk", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	want := []search.Item{{ID: "b-2", Title: "milk shown"}, {ID: "c-3", Title: "milk also shown"}}
	if diff := cmp.Diff(want, out.Notes); diff != "" {
		t.Errorf("Notes (-want +got):\n%s", diff)
	}
}
