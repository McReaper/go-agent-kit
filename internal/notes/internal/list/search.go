package list

import (
	"strings"

	"example.com/notes/internal/notes/internal/domain"
)

func Search(notes []domain.Note, q string) []Item {
	var out []Item
	needle := strings.ToLower(q)
	for _, n := range notes {
		if strings.Contains(strings.ToLower(string(n.Title)), needle) {
			out = append(out, Item{ID: string(n.ID), Title: string(n.Title)})
		}
	}
	return out
}
