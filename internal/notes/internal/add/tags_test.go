package add_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/add"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
)

func TestAddSavesTheTagsLowercasedAndWithoutRepeats(t *testing.T) {
	store := &memStore{}
	_, err := add.New(store, fixedClock).Run(context.Background(), add.Args{Title: "x", Tags: []string{"Go", "go", "Rust"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.saved) != 1 || !slices.Equal(store.saved[0].Tags, []domain.Tag{"go", "rust"}) {
		t.Errorf("saved %+v, want the tags go and rust", store.saved)
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
