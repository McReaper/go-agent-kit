package search

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
)

type Args struct {
	Text  string `json:"text" jsonschema:"the text to find in each note's title or text, ignoring case"`
	Limit int    `json:"limit,omitempty" jsonschema:"how many notes to return, newest first: 20 by default, 100 at most"`
}

type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Out struct {
	Notes []Item `json:"notes"`
}

type Reader interface {
	All(ctx context.Context) ([]domain.Note, error)
}

type Search struct {
	store Reader
}

func New(store Reader) *Search {
	return &Search{store: store}
}

func (s *Search) Op() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "search_notes",
		Description: "Finds the notes whose title or text contains the given text, ignoring case. Returns their ids and titles, newest first.",
		Effect:      op.ReadOnly,
		Run:         s.Run,
	}
}

var errNoText = errors.New("give the text to search for")

// WHY: an empty text would match every note, which is list_notes' job.
func (s *Search) Run(ctx context.Context, args Args) (Out, error) {
	text := strings.TrimSpace(args.Text)
	if text == "" {
		return Out{}, fmt.Errorf("search the notes: %w", errNoText)
	}
	limit, err := domain.CheckLimit(args.Limit)
	if err != nil {
		return Out{}, fmt.Errorf("search the notes: %w", err)
	}
	notes, err := s.store.All(ctx)
	if err != nil {
		return Out{}, fmt.Errorf("search the notes: %w", err)
	}
	out := Out{Notes: []Item{}}
	for _, n := range notes {
		if len(out.Notes) == limit {
			break
		}
		if n.Contains(text) {
			out.Notes = append(out.Notes, Item{ID: string(n.ID), Title: string(n.Title)})
		}
	}
	return out, nil
}
