package list

import (
	"context"
	"fmt"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
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
	limit, err := domain.CheckLimit(args.Limit)
	if err != nil {
		return Out{}, err
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
