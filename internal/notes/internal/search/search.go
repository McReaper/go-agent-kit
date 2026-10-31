package search

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

type Args struct {
	Text string `json:"text" jsonschema:"the text to look for in each note's title and body, ignoring case"`
}

type Item struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags,omitempty"`
}

type Out struct {
	Notes []Item `json:"notes"`
}

type Reader interface {
	Newest(ctx context.Context, limit int) ([]domain.Note, error)
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
		Description: "Returns every note whose title or body contains the given text, ignoring case, newest first.",
		Effect:      op.ReadOnly,
		Run:         s.Run,
	}
}

func (s *Search) Run(ctx context.Context, args Args) (Out, error) {
	if args.Text == "" {
		return Out{}, errors.New("search notes: the text is empty")
	}
	notes, err := s.store.Newest(ctx, math.MaxInt)
	if err != nil {
		return Out{}, fmt.Errorf("search notes: %w", err)
	}
	needle := strings.ToLower(args.Text)
	out := Out{Notes: make([]Item, 0)}
	for _, n := range notes {
		if strings.Contains(strings.ToLower(string(n.Title)), needle) || strings.Contains(strings.ToLower(n.Body), needle) {
			out.Notes = append(out.Notes, toItem(n))
		}
	}
	return out, nil
}

func toItem(n domain.Note) Item {
	item := Item{ID: string(n.ID), Title: string(n.Title)}
	for _, t := range n.Tags {
		item.Tags = append(item.Tags, string(t))
	}
	return item
}
