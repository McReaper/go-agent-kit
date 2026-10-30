package add_test

import (
	"context"
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/add"
	"example.com/notes/internal/notes/internal/domain"
)

type okStore struct{}

func (okStore) Save(context.Context, domain.Note) error { return nil }

func TestAddRuns(t *testing.T) {
	_, err := add.New(okStore{}, time.Now).Run(t.Context(), add.Args{Title: "x"})
	if err != nil {
		t.Fatal(err)
	}
}
