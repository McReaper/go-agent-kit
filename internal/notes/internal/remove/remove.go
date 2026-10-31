package remove

import (
	"context"
	"fmt"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

type Args struct {
	ID string `json:"id" jsonschema:"the id of the note to delete, as add_note and list_notes return it"`
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
		Description: "Deletes the note with the given id, as add_note and list_notes return it. Refuses an id no note has, and says so.",
		Effect:      op.Destructive,
		Run:         d.Run,
	}
}

func (d *Remove) Run(ctx context.Context, args Args) (Out, error) {
	id, err := domain.ParseNoteID(args.ID)
	if err != nil {
		return Out{}, fmt.Errorf("delete a note: %w", err)
	}
	if err := d.store.Delete(ctx, id); err != nil {
		return Out{}, fmt.Errorf("delete a note: %w", err)
	}
	return Out{ID: string(id)}, nil
}
