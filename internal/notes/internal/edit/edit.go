package edit

import (
	"context"
	"errors"
	"fmt"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
)

// WHY: the id is kept on edit. Agents hold ids from add_note and list_notes,
// so a retitled note must still answer to the id they have.
type Args struct {
	ID        string `json:"id" jsonschema:"the note's id, as add_note and list_notes return it"`
	Title     string `json:"title,omitempty" jsonschema:"the new title, at most 80 characters; leave it out to keep the title"`
	Body      string `json:"body,omitempty" jsonschema:"the new text; leave it out to keep the text"`
	ClearBody bool   `json:"clear_body,omitempty" jsonschema:"empty the text; use it instead of body to remove the text"`
}

type Out struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Store interface {
	Get(ctx context.Context, id domain.NoteID) (domain.Note, error)
	Replace(ctx context.Context, n domain.Note) error
}

type Edit struct {
	store Store
}

func New(store Store) *Edit {
	return &Edit{store: store}
}

func (e *Edit) Op() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "edit_note",
		Description: "Changes the title and/or the text of a note, found by its id. The id and the note's date stay the same.",
		Effect:      op.Destructive,
		Run:         e.Run,
	}
}

var (
	errNoID       = errors.New("give the id of the note to edit")
	errNothing    = errors.New("nothing to change: give a title, a body or clear_body")
	errBodyAndClr = errors.New("give body or clear_body, not both")
)

func (e *Edit) Run(ctx context.Context, args Args) (Out, error) {
	if err := checkArgs(args); err != nil {
		return Out{}, fmt.Errorf("edit a note: %w", err)
	}
	title, err := parseTitle(args.Title)
	if err != nil {
		return Out{}, fmt.Errorf("edit a note: %w", err)
	}
	n, err := e.store.Get(ctx, domain.NoteID(args.ID))
	if err != nil {
		return Out{}, fmt.Errorf("edit a note: %w", err)
	}
	apply(&n, args, title)
	if err := e.store.Replace(ctx, n); err != nil {
		return Out{}, fmt.Errorf("edit a note: %w", err)
	}
	return Out{ID: string(n.ID), Title: string(n.Title)}, nil
}

func checkArgs(args Args) error {
	switch {
	case args.ID == "":
		return errNoID
	case args.Body != "" && args.ClearBody:
		return errBodyAndClr
	case args.Title == "" && args.Body == "" && !args.ClearBody:
		return errNothing
	}
	return nil
}

func parseTitle(s string) (domain.Title, error) {
	if s == "" {
		return "", nil
	}
	return domain.ParseTitle(s)
}

// WHY: an empty title means keep the title, which is safe because an empty
// title is never valid, so no note can have one.
func apply(n *domain.Note, args Args, title domain.Title) {
	if title != "" {
		n.Title = title
	}
	if args.Body != "" {
		n.Body = args.Body
	}
	if args.ClearBody {
		n.Body = ""
	}
}
