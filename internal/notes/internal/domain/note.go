package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/McReaper/go-agent-kit/template/pkg/slug"
)

const (
	MaxTitleRunes = 80
	MaxTagRunes   = 32
)

var (
	ErrEmptyTitle    = errors.New("the title is empty")
	ErrNoSlugInTitle = errors.New("the title has no letter or digit to name the note by")
	ErrBadID         = errors.New("a note id is lowercase letters and digits joined by single hyphens")
	ErrEmptyTag      = errors.New("a tag is empty")
	ErrNoSuchNote    = errors.New("no note has this id")
)

var idShape = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type NoteID string

func ParseNoteID(s string) (NoteID, error) {
	if !idShape.MatchString(s) {
		return "", fmt.Errorf("note id %q: %w", s, ErrBadID)
	}
	return NoteID(s), nil
}

type Title string

// Tag is a note's tag. Tags are case-insensitive, so a Tag is always lowercase.
type Tag string

type Note struct {
	ID      NoteID
	Title   Title
	Body    string
	Tags    []Tag
	Created time.Time
}

// HasTag reports whether n carries t.
func (n Note) HasTag(t Tag) bool {
	return slices.Contains(n.Tags, t)
}

// ParseTag trims and lowercases s into a tag.
func ParseTag(s string) (Tag, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return "", ErrEmptyTag
	}
	if n := utf8.RuneCountInString(t); n > MaxTagRunes {
		return "", fmt.Errorf("the tag %q has %d characters, the limit is %d", t, n, MaxTagRunes)
	}
	return Tag(t), nil
}

// ParseTags parses each of raw into a tag, dropping repeats. It returns nil
// when raw has no tags.
func ParseTags(raw []string) ([]Tag, error) {
	var tags []Tag
	for _, s := range raw {
		t, err := ParseTag(s)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(tags, t) {
			tags = append(tags, t)
		}
	}
	return tags, nil
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

func New(title Title, body string, at time.Time) Note {
	at = at.UTC()
	id := NoteID(slug.Make(string(title)) + "-" + at.Format("20060102t150405"))
	return Note{ID: id, Title: title, Body: body, Created: at}
}
