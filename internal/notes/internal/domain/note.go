package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"example.com/notes/pkg/slug"
)

const MaxTitleRunes = 80

var (
	ErrEmptyTitle    = errors.New("the title is empty")
	ErrNoSlugInTitle = errors.New("the title has no letter or digit to name the note by")
	ErrEmptyID       = errors.New("the id is empty")
	ErrBadID         = errors.New("the id has characters other than a-z, 0-9 and hyphen")
)

type NoteID string

type Title string

type Note struct {
	ID      NoteID
	Title   Title
	Body    string
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

// WHY: an id names a file, so only the characters slug.Make produces, and
// the timestamp New appends, are allowed.
func ParseID(s string) (NoteID, error) {
	if s == "" {
		return "", ErrEmptyID
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return "", ErrBadID
		}
	}
	return NoteID(s), nil
}

func New(title Title, body string, at time.Time) Note {
	at = at.UTC()
	id := NoteID(slug.Make(string(title)) + "-" + at.Format("20060102t150405"))
	return Note{ID: id, Title: title, Body: body, Created: at}
}
