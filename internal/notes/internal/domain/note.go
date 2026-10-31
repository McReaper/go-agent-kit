package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/McReaper/go-agent-kit/template/pkg/slug"
)

const MaxTitleRunes = 80

var (
	ErrEmptyTitle    = errors.New("the title is empty")
	ErrNoSlugInTitle = errors.New("the title has no letter or digit to name the note by")
	ErrBadID         = errors.New("a note id is lowercase letters and digits joined by single hyphens")
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

func New(title Title, body string, at time.Time) Note {
	at = at.UTC()
	id := NoteID(slug.Make(string(title)) + "-" + at.Format("20060102t150405"))
	return Note{ID: id, Title: title, Body: body, Created: at}
}
