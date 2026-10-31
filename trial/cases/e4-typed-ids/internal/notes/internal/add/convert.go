package add

import "example.com/notes/internal/notes/internal/domain"

func FromString(s string) domain.NoteID {
	return domain.NoteID(s)
}
