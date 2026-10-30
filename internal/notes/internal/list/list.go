package list

import (
	"context"
	"fmt"
	"math"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/platform/op"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Args struct {
	Limit int    `json:"limit,omitempty" jsonschema:"how many notes to return, newest first: 20 by default, 100 at most"`
	Tag   string `json:"tag,omitempty" jsonschema:"only the notes with this tag"`
}

type Item struct {
	ID    string       `json:"id"`
	Title string       `json:"title"`
	Tags  []domain.Tag `json:"tags,omitempty"`
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
		Description: "Lists the newest notes, their ids, titles and tags, or only the notes carrying one tag.",
		Effect:      op.ReadOnly,
		Run:         l.Run,
	}
}

func (l *List) Run(ctx context.Context, args Args) (Out, error) {
	limit, err := pageLimit(args.Limit)
	if err != nil {
		return Out{}, err
	}
	var tag domain.Tag
	if args.Tag != "" {
		if tag, err = domain.ParseTag(args.Tag); err != nil {
			return Out{}, fmt.Errorf("list the notes: %w", err)
		}
	}
	// WHY: the store pages only the newest notes, so a tag filter reads them
	// all and takes the limit after the filter; cutting first would return
	// fewer notes than asked while older ones carry the tag.
	fetch := limit
	if tag != "" {
		fetch = math.MaxInt
	}
	notes, err := l.store.Newest(ctx, fetch)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	return Out{Notes: items(notes, tag, limit)}, nil
}

func pageLimit(limit int) (int, error) {
	switch {
	case limit == 0:
		return DefaultLimit, nil
	case limit < 0 || limit > MaxLimit:
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, MaxLimit)
	}
	return limit, nil
}

func items(notes []domain.Note, tag domain.Tag, limit int) []Item {
	out := make([]Item, 0, min(limit, len(notes)))
	for _, n := range notes {
		if len(out) == limit {
			break
		}
		if tag != "" && !n.HasTag(tag) {
			continue
		}
		out = append(out, Item{ID: string(n.ID), Title: string(n.Title), Tags: n.Tags})
	}
	return out
}
