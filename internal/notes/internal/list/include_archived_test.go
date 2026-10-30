package list_test

import (
	"context"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/list"
)

type flagStore struct {
	includeArchived bool
}

func (s *flagStore) Recent(_ context.Context, _ int, includeArchived bool) ([]domain.Note, error) {
	s.includeArchived = includeArchived
	return nil, nil
}

func TestListPassesTheIncludeArchivedChoiceToTheStore(t *testing.T) {
	for _, include := range []bool{false, true} {
		store := &flagStore{}
		if _, err := list.New(store).Run(context.Background(), list.Args{IncludeArchived: include}); err != nil {
			t.Fatal(err)
		}
		if store.includeArchived != include {
			t.Errorf("IncludeArchived %v reached the store as %v", include, store.includeArchived)
		}
	}
}
