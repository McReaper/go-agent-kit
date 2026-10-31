package list

import (
	"context"
	"fmt"
	"math"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Args struct {
	Limit *int    `json:"limit,omitempty" jsonschema:"how many notes to return, newest first: 1 to 100, 20 if not given"`
	Tag   *string `json:"tag,omitempty" jsonschema:"only the notes carrying this tag, case-insensitive; every note if not given"`
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

type List struct {
	store Reader
}

func New(store Reader) *List {
	return &List{store: store}
}

func (l *List) Op() op.Op[Args, Out] {
	return op.Op[Args, Out]{
		Name:        "list_notes",
		Description: "Lists the newest notes, their ids, titles and tags, or only those carrying a given tag.",
		Effect:      op.ReadOnly,
		Run:         l.Run,
	}
}

func (l *List) Run(ctx context.Context, args Args) (Out, error) {
	limit := defaultLimit
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit < 1 || limit > maxLimit {
		return Out{}, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, maxLimit)
	}
	if args.Tag != nil {
		return l.newestWithTag(ctx, *args.Tag, limit)
	}
	notes, err := l.store.Newest(ctx, limit)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	return toOut(notes), nil
}

// WHY: the store cuts to the newest notes before a tag can narrow them, so a
// tag reads every note and keeps the first limit of those that carry it.
func (l *List) newestWithTag(ctx context.Context, raw string, limit int) (Out, error) {
	tag, err := domain.ParseTag(raw)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	notes, err := l.store.Newest(ctx, math.MaxInt)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	var tagged []domain.Note
	for _, n := range notes {
		if n.HasTag(tag) {
			tagged = append(tagged, n)
		}
	}
	return toOut(tagged[:min(limit, len(tagged))]), nil
}

func toOut(notes []domain.Note) Out {
	out := Out{Notes: make([]Item, 0, len(notes))}
	for _, n := range notes {
		out.Notes = append(out.Notes, Item{ID: string(n.ID), Title: string(n.Title), Tags: tagStrings(n.Tags)})
	}
	return out
}

// WHY: a note without tags has no tags key, so the list stays as it was for
// notes that were never tagged.
func tagStrings(tags []domain.Tag) []string {
	var out []string
	for _, t := range tags {
		out = append(out, string(t))
	}
	return out
}
