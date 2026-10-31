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
	Tag   *string `json:"tag,omitempty" jsonschema:"list only the notes filed under this tag, case-insensitive; the limit counts those notes"`
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
		Description: "Lists the newest notes, their ids, titles and tags. Given a tag, lists only the notes filed under it.",
		Effect:      op.ReadOnly,
		Run:         l.Run,
	}
}

func (l *List) Run(ctx context.Context, args Args) (Out, error) {
	limit, err := limitOf(args.Limit)
	if err != nil {
		return Out{}, err
	}
	tag, filter, err := tagOf(args.Tag)
	if err != nil {
		return Out{}, err
	}
	// WHY: the store returns the newest first, and a tag filter needs every
	// note, so the limit counts the notes that match rather than the ones read.
	want := limit
	if filter {
		want = math.MaxInt
	}
	notes, err := l.store.Newest(ctx, want)
	if err != nil {
		return Out{}, fmt.Errorf("list the notes: %w", err)
	}
	return Out{Notes: pick(notes, limit, tag, filter)}, nil
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

func tagOf(p *string) (domain.Tag, bool, error) {
	if p == nil {
		return "", false, nil
	}
	t, err := domain.ParseTag(*p)
	if err != nil {
		return "", false, fmt.Errorf("list the notes: %w", err)
	}
	return t, true, nil
}

func pick(notes []domain.Note, limit int, tag domain.Tag, filter bool) []Item {
	items := make([]Item, 0, min(limit, len(notes)))
	for _, n := range notes {
		if len(items) < limit && (!filter || n.HasTag(tag)) {
			items = append(items, item(n))
		}
	}
	return items
}

func item(n domain.Note) Item {
	it := Item{ID: string(n.ID), Title: string(n.Title)}
	for _, t := range n.Tags {
		it.Tags = append(it.Tags, string(t))
	}
	return it
}
