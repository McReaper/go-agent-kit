package add

import "github.com/McReaper/rumblestrip-go/template/internal/notes/internal/domain"

func FromString(s string) domain.NoteID {
	return domain.NoteID(s)
}
