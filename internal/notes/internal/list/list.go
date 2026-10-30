package list

import (
	"context"
	"fmt"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
)

type Args struct {
	Limit           int  `json:"limit,omitempty" jsonschema:"how many notes to return, newest first: 20 by default, 100 at most"`
	IncludeArchived bool `json:"include_archived,omitempty" jsonschema:"also list the archived notes, which are left out by default"`
}

type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Out struct {
	Notes []Item `json:"notes"`
}

type Reader interface {
	Recent(ctx context.Context, limit int, includeArchived bool) ([]domain.Note, error)
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
		Description: "Lists the newest notes, their ids and titles. Archived notes are left out unless include_archived is set.",
		Effect:      op.ReadOnly,
		Run:         l.Run,
	}
}

func (l *List) Run(ctx context.Context, args Args) (Out, error) {
	limit, err := domain.CheckLimit(args.Limit)
	if err != nil {
		return Out{}, err
	}
	notes, err := l.store.Recent(ctx, limit, args.IncludeArchived)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	out := Out{Notes: make([]Item, 0, len(notes))}
	for _, n := range notes {
		out.Notes = append(out.Notes, Item{ID: string(n.ID), Title: string(n.Title)})
	}
	return out, nil
}
