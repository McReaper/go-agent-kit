package domain

import "errors"

type NoteID string

type Note struct {
	ID    NoteID
	Count int
}

func ParseNoteID(s string) (NoteID, error) {
	if s == "" {
		return "", errors.New("empty id")
	}
	return NoteID(s), nil
}
