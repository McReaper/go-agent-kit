package add_test

import (
	"context"
	"errors"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/add"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/google/go-cmp/cmp"
)

func TestAddSavesNormalizedTags(t *testing.T) {
	store := &memStore{}
	if _, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: "x", Tags: []string{"Go", "go", "Tools"}}); err != nil {
		t.Fatal(err)
	}
	if len(store.saved) != 1 {
		t.Fatalf("saved %d notes", len(store.saved))
	}
	if diff := cmp.Diff([]domain.Tag{"go", "tools"}, store.saved[0].Tags); diff != "" {
		t.Errorf("tags (-want +got):\n%s", diff)
	}
}

func TestAddRefusesAnInvalidTagWithoutSaving(t *testing.T) {
	store := &memStore{}
	_, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: "x", Tags: []string{"  "}})
	if !errors.Is(err, domain.ErrEmptyTag) {
		t.Fatalf("err = %v, want ErrEmptyTag", err)
	}
	if len(store.saved) != 0 {
		t.Errorf("saved %d notes", len(store.saved))
	}
}
