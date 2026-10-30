package domain_test

import (
	"testing"

	"example.com/notes/internal/notes/internal/domain"
)

func TestVisibleHidesArchivedNotesUnlessAskedFor(t *testing.T) {
	live := domain.Note{ID: "a-1"}
	archived := domain.Note{ID: "b-2", Archived: true}
	cases := []struct {
		name     string
		note     domain.Note
		include  bool
		wantSeen bool
	}{
		{"a live note is shown", live, false, true},
		{"a live note is shown when archived ones are asked for", live, true, true},
		{"an archived note is hidden", archived, false, false},
		{"an archived note is shown when asked for", archived, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.note.Visible(c.include); got != c.wantSeen {
				t.Errorf("Visible(%v) = %v, want %v", c.include, got, c.wantSeen)
			}
		})
	}
}
