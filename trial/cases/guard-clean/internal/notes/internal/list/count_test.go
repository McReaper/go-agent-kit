package list_test

import (
	"testing"

	"example.com/notes/internal/notes/internal/list"
)

func TestCountCountsTheNotes(t *testing.T) {
	if n := list.Count(list.Out{Notes: make([]list.Item, 3)}); n != 3 {
		t.Errorf("Count = %d", n)
	}
}
