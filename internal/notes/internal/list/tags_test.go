package list_test

import (
	"context"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/list"
	"github.com/google/go-cmp/cmp"
)

func TestListReturnsTheTagsOfEachNote(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "a-1", Title: "A", Tags: []domain.Tag{"go", "todo"}}}}
	out, err := list.New(store).Run(context.Background(), list.Args{})
	if err != nil {
		t.Fatal(err)
	}
	want := list.Out{Notes: []list.Item{{ID: "a-1", Title: "A", Tags: []string{"go", "todo"}}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestListReturnsOnlyTheNotesWithTheTagIgnoringCase(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "A", Tags: []domain.Tag{"go"}},
		{ID: "b-2", Title: "B"},
		{ID: "c-3", Title: "C", Tags: []domain.Tag{"rust", "go"}},
	}}
	out, err := list.New(store).Run(context.Background(), list.Args{Tag: new("GO")})
	if err != nil {
		t.Fatal(err)
	}
	want := list.Out{Notes: []list.Item{
		{ID: "a-1", Title: "A", Tags: []string{"go"}},
		{ID: "c-3", Title: "C", Tags: []string{"rust", "go"}},
	}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestListCountsTheLimitAfterTheTagFilter(t *testing.T) {
	store := &memStore{notes: []domain.Note{
		{ID: "a-1", Title: "A"},
		{ID: "b-2", Title: "B"},
		{ID: "c-3", Title: "C", Tags: []domain.Tag{"go"}},
	}}
	out, err := list.New(store).Run(context.Background(), list.Args{Tag: new("go"), Limit: new(1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 1 || out.Notes[0].ID != "c-3" {
		t.Errorf("Run = %+v, want the one tagged note c-3", out.Notes)
	}
}

func TestListRefusesAnEmptyTagFilter(t *testing.T) {
	if _, err := list.New(&memStore{}).Run(context.Background(), list.Args{Tag: new("  ")}); err == nil {
		t.Error("an empty tag filter was accepted")
	}
}
