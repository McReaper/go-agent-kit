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

const (
	defaultLimit = 20
	maxLimit     = 100
)

// WHY: a blank text would match every note, which is never what was asked.
var ErrEmptyText = errors.New("the text to search for is empty")

type Args struct {
	Text  string `json:"text" jsonschema:"the text to find: a note matches when its title or its body contains it, ignoring case"`
	Limit *int   `json:"limit,omitempty" jsonschema:"how many matching notes to return, newest first: 1 to 100, 20 if not given"`
}

type Item struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Body  string   `json:"body,omitempty"`
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
		Description: "Lists the notes whose title or body contains the text given, ignoring case, newest first. Refuses an empty text.",
		Effect:      op.ReadOnly,
		Run:         s.Run,
	}
}

func (s *Search) Run(ctx context.Context, args Args) (Out, error) {
	if strings.TrimSpace(args.Text) == "" {
		return Out{}, fmt.Errorf("search the notes: %w", ErrEmptyText)
	}
	limit, err := limitOf(args.Limit)
	if err != nil {
		return Out{}, err
	}
	// WHY: a match can be any note, so every note is read, and the limit
	// counts the matches rather than the notes read.
	notes, err := s.store.Newest(ctx, math.MaxInt)
	if err != nil {
		return Out{}, fmt.Errorf("search the notes: %w", err)
	}
	return Out{Notes: pick(notes, strings.ToLower(args.Text), limit)}, nil
}

func limitOf(p *int) (int, error) {
	limit := defaultLimit
	if p != nil {
		limit = *p
	}
	if limit < 1 || limit > maxLimit {
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, maxLimit)
	}
	return limit, nil
}

func pick(notes []domain.Note, needle string, limit int) []Item {
	items := []Item{}
	for _, n := range notes {
		if matches(n, needle) && len(items) < limit {
			items = append(items, item(n))
		}
	}
	return items
}

func matches(n domain.Note, needle string) bool {
	return strings.Contains(strings.ToLower(string(n.Title)), needle) ||
		strings.Contains(strings.ToLower(n.Body), needle)
}

func item(n domain.Note) Item {
	it := Item{ID: string(n.ID), Title: string(n.Title), Body: n.Body}
	for _, t := range n.Tags {
		it.Tags = append(it.Tags, string(t))
	}
	return it
}
