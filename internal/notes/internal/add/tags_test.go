package add_test

import (
	"context"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"github.com/google/go-cmp/cmp"

	"example.com/notes/internal/notes/internal/add"
)

func TestAddSavesTheTags(t *testing.T) {
	store := &memStore{}
	if _, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: "Milk", Tags: []string{"groceries", "2026"}}); err != nil {
		t.Fatal(err)
	}
	if len(store.saved) != 1 {
		t.Fatalf("saved %d notes", len(store.saved))
	}
	want := []domain.Tag{"groceries", "2026"}
	if diff := cmp.Diff(want, store.saved[0].Tags); diff != "" {
		t.Errorf("tags (-want +got):\n%s", diff)
	}
}

func TestAddRefusesABadTagWithoutSaving(t *testing.T) {
	for _, tags := range [][]string{{"Work"}, {"a", "b", "c", "d", "e", "f"}, {"x", "x"}} {
		store := &memStore{}
		_, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: "Milk", Tags: tags})
		if err == nil {
			t.Errorf("tags %v accepted", tags)
		}
		if len(store.saved) != 0 {
			t.Errorf("tags %v: saved %d notes", tags, len(store.saved))
		}
	}
}
