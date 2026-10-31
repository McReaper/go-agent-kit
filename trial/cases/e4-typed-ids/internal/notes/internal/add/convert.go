package add

import "github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"

func FromString(s string) domain.NoteID {
	return domain.NoteID(s)
}
