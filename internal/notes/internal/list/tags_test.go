package list_test

import (
	"context"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/list"
	"github.com/google/go-cmp/cmp"
)

func tagged(id string, tags ...domain.Tag) domain.Note {
	return domain.Note{ID: domain.NoteID(id), Title: domain.Title(id), Tags: tags}
}

func TestListFiltersByTagIgnoringCase(t *testing.T) {
	store := &memStore{notes: []domain.Note{tagged("a", "go"), tagged("b", "rust"), tagged("c", "go", "rust")}}
	tag := "GO"
	out, err := list.New(store).Run(context.Background(), list.Args{Tag: &tag})
	if err != nil {
		t.Fatal(err)
	}
	want := list.Out{Notes: []list.Item{
		{ID: "a", Title: "a", Tags: []string{"go"}},
		{ID: "c", Title: "c", Tags: []string{"go", "rust"}},
	}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestListAppliesTheLimitAfterTheTagFilter(t *testing.T) {
	store := &memStore{notes: []domain.Note{tagged("a"), tagged("b", "go"), tagged("c"), tagged("d", "go")}}
	tag := "go"
	out, err := list.New(store).Run(context.Background(), list.Args{Tag: &tag, Limit: new(1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 1 || out.Notes[0].ID != "b" {
		t.Errorf("Run = %+v, want the first tagged note only", out.Notes)
	}
}

func TestListRefusesAnEmptyTag(t *testing.T) {
	tag := "  "
	if _, err := list.New(&memStore{}).Run(context.Background(), list.Args{Tag: &tag}); err == nil {
		t.Error("an empty tag filter was accepted")
	}
}
