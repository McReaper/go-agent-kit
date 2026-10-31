package list

import (
	"context"
	"fmt"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Args struct {
	Limit int `json:"limit,omitempty" jsonschema:"how many notes to return, newest first: 20 by default, 100 at most"`
}

type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Out struct {
	Notes []Item `json:"notes"`
}

type Reader interface {
	Newest(ctx context.Context, limit int) ([]domain.Note, error)
}

type List struct {
	store Reader
}

func New(store Reader) *List {
	return &List{store: store}
}

func (l *List) Op() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "list_notes",
		Description: "Lists the newest notes, their ids and titles.",
		Effect:      op.ReadOnly,
		Run:         l.Run,
	}
}

func (l *List) Run(ctx context.Context, args Args) (Out, error) {
	limit := args.Limit
	switch {
	case limit == 0:
		limit = defaultLimit
	case limit < 0 || limit > maxLimit:
		return Out{}, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, maxLimit)
	}
	notes, err := l.store.Newest(ctx, limit)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	out := Out{Notes: make([]Item, 0, len(notes))}
	for _, n := range notes {
		out.Notes = append(out.Notes, Item{ID: string(n.ID), Title: string(n.Title)})
	}
	return out, nil
}
