package list

import (
	"strings"

	"example.com/notes/internal/notes/internal/domain"
)

// Search returns the notes whose title contains q.
func Search(notes []domain.Note, q string) []Item {
	var out []Item
	for _, n := range notes {
		for _, w := range strings.Fields(string(n.Title)) {
			if strings.EqualFold(w, q) {
				out = append(out, Item{ID: string(n.ID), Title: string(n.Title)})
			}
		}
	}
	return out
}
