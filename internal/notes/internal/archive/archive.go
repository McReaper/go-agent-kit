package archive

import (
	"context"
	"errors"
	"fmt"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
)

// WHY: both operations take only the id, and share one rule, so an archived
// note comes back exactly as it was when unarchived.
type Args struct {
	ID string `json:"id" jsonschema:"the note's id, as add_note and list_notes return it"`
}

type Out struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Store interface {
	Get(ctx context.Context, id domain.NoteID) (domain.Note, error)
	Replace(ctx context.Context, n domain.Note) error
}

type Archive struct {
	store Store
}

func New(store Store) *Archive {
	return &Archive{store: store}
}

func (a *Archive) ArchiveOp() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "archive_note",
		Description: "Archives a note, found by its id: it no longer appears in list_notes or search_notes unless they include archived notes. Nothing is deleted; unarchive_note shows it again.",
		Effect:      op.Destructive,
		Run: func(ctx context.Context, args Args) (Out, error) {
			return a.set(ctx, "archive", args, true)
		},
	}
}

func (a *Archive) UnarchiveOp() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "unarchive_note",
		Description: "Unarchives a note, found by its id, so list_notes and search_notes show it again.",
		Effect:      op.Destructive,
		Run: func(ctx context.Context, args Args) (Out, error) {
			return a.set(ctx, "unarchive", args, false)
		},
	}
}

var errNoID = errors.New("give the id of the note")

// WHY: archiving a note that is already archived succeeds without change, so
// an agent that repeats a call after a timeout is not told it failed.
func (a *Archive) set(ctx context.Context, what string, args Args, archived bool) (Out, error) {
	if args.ID == "" {
		return Out{}, fmt.Errorf("%s a note: %w", what, errNoID)
	}
	n, err := a.store.Get(ctx, domain.NoteID(args.ID))
	if err != nil {
		return Out{}, fmt.Errorf("%s a note: %w", what, err)
	}
	n.Archived = archived
	if err := a.store.Replace(ctx, n); err != nil {
		return Out{}, fmt.Errorf("%s a note: %w", what, err)
	}
	return Out{ID: string(n.ID), Title: string(n.Title)}, nil
}
