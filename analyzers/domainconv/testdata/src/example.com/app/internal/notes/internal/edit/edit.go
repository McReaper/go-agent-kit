package edit

import "example.com/app/internal/notes/internal/domain"

func Bad(s string) domain.NoteID {
	return domain.NoteID(s) // want `E4: NoteID is built by conversion outside the domain: call the domain's ParseNoteID`
}

func Good(s string) (domain.NoteID, error) {
	return domain.ParseNoteID(s)
}

func Back(id domain.NoteID) string {
	return string(id)
}

func Struct(id domain.NoteID) domain.Note {
	return domain.Note(domain.Note{ID: id})
}
