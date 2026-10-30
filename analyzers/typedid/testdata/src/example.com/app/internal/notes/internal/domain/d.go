package domain

type NoteID string

type Note struct {
	ID     NoteID
	UserID string // want `E4: UserID is a bare string`
	Ref    int
}

func Find(id string, ownerId int64) (parentID NoteID) { // want `E4: id is a bare string` `E4: ownerId is a bare int64`
	return NoteID(id)
}

func Valid(identity string, kids int) bool {
	return identity != "" && kids >= 0
}
