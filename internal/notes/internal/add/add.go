package add

import (
	"context"
	"fmt"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

type Args struct {
	Title string `json:"title" jsonschema:"the note's title, at most 80 characters"`
	Body  string `json:"body,omitempty" jsonschema:"the note's text"`
}

type Out struct {
	ID string `json:"id"`
}

type Saver interface {
	Save(ctx context.Context, n domain.Note) error
}

type Add struct {
	store Saver
	now   func() time.Time
}

func New(store Saver, now func() time.Time) *Add {
	return &Add{store: store, now: now}
}

func (a *Add) Op() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "add_note",
		Description: "Adds a note and returns its id. Never overwrites: a second note with the same title in the same second is refused.",
		Effect:      op.Additive,
		Run:         a.Run,
	}
}

func (a *Add) Run(ctx context.Context, args Args) (Out, error) {
	title, err := domain.ParseTitle(args.Title)
	if err != nil {
		return Out{}, fmt.Errorf("add a note: %w", err)
	}
	n := domain.New(title, args.Body, a.now())
	if err := a.store.Save(ctx, n); err != nil {
		return Out{}, fmt.Errorf("add a note: %w", err)
	}
	return Out{ID: string(n.ID)}, nil
}
