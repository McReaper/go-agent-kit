package remove

import (
	"context"
	"fmt"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
)

type Args struct {
	ID string `json:"id" jsonschema:"the id of the note to delete, as add_note or list_notes returned it"`
}

type Out struct {
	ID string `json:"id"`
}

type Deleter interface {
	Delete(ctx context.Context, id domain.NoteID) error
}

type Remove struct {
	store Deleter
}

func New(store Deleter) *Remove {
	return &Remove{store: store}
}

func (d *Remove) Op() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "delete_note",
		Description: "Permanently deletes the note with the given id. It can't be undone, and an id with no note is an error.",
		Effect:      op.Destructive,
		Run:         d.Run,
	}
}

func (d *Remove) Run(ctx context.Context, args Args) (Out, error) {
	id, err := domain.ParseID(args.ID)
	if err != nil {
		return Out{}, fmt.Errorf("delete a note: %w", err)
	}
	if err := d.store.Delete(ctx, id); err != nil {
		return Out{}, fmt.Errorf("delete a note: %w", err)
	}
	return Out{ID: string(id)}, nil
}
