package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"example.com/notes/pkg/slug"
)

const (
	MaxTitleRunes = 80
	MaxTags       = 5
)

var (
	ErrEmptyTitle    = errors.New("the title is empty")
	ErrNoSlugInTitle = errors.New("the title has no letter or digit to name the note by")
	ErrEmptyTag      = errors.New("a tag is empty")
)

type NoteID string

type Title string

type Tag string

type Note struct {
	ID      NoteID
	Title   Title
	Body    string
	Tags    []Tag
	Created time.Time
}

func ParseTitle(s string) (Title, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return "", ErrEmptyTitle
	}
	if n := utf8.RuneCountInString(t); n > MaxTitleRunes {
		return "", fmt.Errorf("the title has %d characters, the limit is %d", n, MaxTitleRunes)
	}
	if slug.Make(t) == "" {
		return "", ErrNoSlugInTitle
	}
	return Title(t), nil
}

// WHY: a tag is one word in the form a filter can match exactly, so a
// spelling variant such as "Work" or "work " is refused, not folded.
func ParseTag(s string) (Tag, error) {
	if s == "" {
		return "", ErrEmptyTag
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return "", fmt.Errorf("the tag %q has %q: use lowercase letters and digits only", s, r)
		}
	}
	return Tag(s), nil
}

func ParseTags(raw []string) ([]Tag, error) {
	if len(raw) > MaxTags {
		return nil, fmt.Errorf("a note has at most %d tags, got %d", MaxTags, len(raw))
	}
	tags := make([]Tag, 0, len(raw))
	for _, s := range raw {
		t, err := ParseTag(s)
		if err != nil {
			return nil, err
		}
		if slices.Contains(tags, t) {
			return nil, fmt.Errorf("the tag %q is given twice", t)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (n Note) HasTag(t Tag) bool {
	return slices.Contains(n.Tags, t)
}

func New(title Title, body string, at time.Time) Note {
	at = at.UTC()
	id := NoteID(slug.Make(string(title)) + "-" + at.Format("20060102t150405"))
	return Note{ID: id, Title: title, Body: body, Created: at}
}
