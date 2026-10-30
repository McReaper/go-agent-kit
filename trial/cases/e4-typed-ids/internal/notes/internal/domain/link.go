package domain

type Link struct {
	From   NoteID
	ToID   string
	Weight int
}

func Linked(from NoteID, userID int64) Link {
	return Link{From: from, Weight: int(userID)}
}
