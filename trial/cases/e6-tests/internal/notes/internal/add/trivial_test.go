package add_test

import (
	"context"
	"testing"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/add"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
)

type okStore struct{}

func (okStore) Save(context.Context, domain.Note) error { return nil }

func TestAddRuns(t *testing.T) {
	_, err := add.New(okStore{}, time.Now).Run(t.Context(), add.Args{Title: "x"})
	if err != nil {
		t.Fatal(err)
	}
}
